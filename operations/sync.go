package operations

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"go.beyondstorage.io/v5/types"
)

func (do *DualOperator) SyncDir(src, dst string, opts SyncOptions) (errch chan *EmptyResult, err error) {
	errch = make(chan *EmptyResult, 4)

	so := NewSingleOperator(do.src)

	var ch chan *ObjectResult
	if opts.Recursive {
		ch, err = so.ListRecursively(src)
	} else {
		ch, err = so.List(src)
	}
	if err != nil {
		return nil, err
	}

	_, err = do.dst.Write(dst, nil, 0)
	if err != nil {
		return nil, err
	}

	filesName, err := getFilesName(dst, do.dst, opts.Recursive)
	if err != nil {
		return nil, err
	}

	var ex, in *regexp.Regexp
	if opts.IsExclude {
		ex = regexp.MustCompile(opts.Exclude)
	}
	if opts.IsInclude {
		in = regexp.MustCompile(opts.Include)
	}

	go func() {
		defer close(errch)

		wg := &sync.WaitGroup{}

		for v := range ch {
			if v.Error != nil {
				errch <- &EmptyResult{Error: v.Error}
				return
			}

			o := v.Object
			if !opts.Recursive && o.Mode.IsDir() {
				continue
			}

			objRelPath := strings.TrimPrefix(o.Path, src)
			if value, ok := filesName[objRelPath]; ok {
				delete(filesName, objRelPath)
				if opts.IgnoreExisting {
					continue
				}
				if !o.Mode.IsDir() && opts.Update && o.MustGetLastModified().Before(value) {
					continue
				}
			} else {
				if opts.Existing {
					continue
				}
			}

			if opts.IsExclude && ex.MatchString(objRelPath) {
				if !opts.IsInclude || !in.MatchString(objRelPath) {
					continue
				}
			}

			wg.Add(1)

			err = do.pool.Submit(func() {
				defer wg.Done()

				var buf bytes.Buffer
				_, err = do.src.Read(o.Path, &buf)
				if err != nil {
					errch <- &EmptyResult{Error: err}
					return
				}

				path := dst + objRelPath
				if opts.IsArgs {
					path = dst + o.Path
				}

				size := int64(buf.Len())
				if size > opts.MultipartThreshold {
					mch, err := do.writeFileViaMultipart(&buf, path, size)
					if err != nil {
						errch <- &EmptyResult{Error: err}
						return
					}

					for value := range mch {
						if value.Error != nil {
							errch <- &EmptyResult{Error: value.Error}
							return
						}
					}
				} else {
					_, err = do.dst.Write(path, &buf, size)
					if err != nil {
						errch <- &EmptyResult{Error: err}
						return
					}
				}

				if !o.Mode.IsDir() {
					if opts.IsArgs {
						fmt.Printf("<%s> synced.\n", o.Path)
					} else {
						fmt.Printf("<%s> synced.\n", objRelPath)
					}
				}
			})
			if err != nil {
				do.logger.Error("submit task", zap.Error(err))
				break
			}
		}

		wg.Wait()
	}()

	if opts.Remove {
		if filesName != nil {
			dstSo := NewSingleOperator(do.dst)
			for k := range filesName {
				err = dstSo.Delete(dst + k)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return
}

func (do *DualOperator) writeFileViaMultipart(r io.Reader, path string, size int64) (errch chan *EmptyResult, err error) {
	errch = make(chan *EmptyResult, 4)
	partch := make(chan *PartResult, 4)

	multiparter, ok := do.dst.(types.Multiparter)
	if !ok {
		return nil, fmt.Errorf("multiparter")
	}

	mo, err := multiparter.CreateMultipart(path)
	if err != nil {
		return nil, err
	}

	partSize, err := calculatePartSize(do.dst, size)
	if err != nil {
		return nil, err
	}

	go func() {
		// Close partch to inform that all parts have been done.
		defer close(partch)

		wg := &sync.WaitGroup{}
		var index int

		b := make([]byte, partSize)

		for {
			wg.Add(1)

			taskIndex := index

			n, err := io.ReadFull(r, b)
			if err == io.EOF {
				wg.Done()
				break
			}
			if err != nil && n != 0 {
				err = nil
			}
			if err != nil {
				partch <- &PartResult{Error: err}
				return
			}

			rd := bytes.NewReader(b[:n])

			err = do.pool.Submit(func() {
				defer wg.Done()

				_, part, err := multiparter.WriteMultipart(mo, rd, rd.Size(), taskIndex)
				if err != nil {
					partch <- &PartResult{Error: err}
					return
				}
				partch <- &PartResult{Part: part}
			})
			if err != nil {
				do.logger.Error("submit task", zap.Error(err))
				return
			}

			index++
		}

		wg.Wait()
	}()

	defer close(errch)

	parts := make([]*types.Part, 0)
	for v := range partch {
		if v.Error != nil {
			errch <- &EmptyResult{Error: v.Error}
			continue
		}
		parts = append(parts, v.Part)
	}

	sort.SliceStable(parts, func(i, j int) bool {
		return parts[i].Index < parts[j].Index
	})

	err = multiparter.CompleteMultipart(mo, parts)
	if err != nil {
		return nil, err
	}

	return
}

func getFilesName(path string, store types.Storager, recursive bool) (files map[string]time.Time, err error) {
	files = make(map[string]time.Time, 0)

	so := NewSingleOperator(store)

	var ch chan *ObjectResult
	if recursive {
		ch, err = so.ListRecursively(path)
	} else {
		ch, err = so.List(path)
	}
	if err != nil {
		return nil, err
	}

	for v := range ch {
		if v.Error != nil {
			return nil, v.Error
		}

		o := v.Object
		if o.Mode.IsDir() {
			continue
		}

		objRelPath := strings.TrimPrefix(v.Object.Path, path)
		files[objRelPath] = o.MustGetLastModified()
	}

	return
}

type SyncOptions struct {
	Recursive          bool
	MultipartThreshold int64
	Existing           bool
	IgnoreExisting     bool
	Remove             bool
	Update             bool
	IsExclude          bool
	Exclude            string
	IsInclude          bool
	Include            string
	IsArgs             bool
}
