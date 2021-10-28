package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"go.beyondstorage.io/beyond-ctl/operations"
	"go.beyondstorage.io/v5/services"
	"go.beyondstorage.io/v5/types"
)

const (
	statFlagJson = "json"
)

var statFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:  statFlagJson,
		Usage: "Output in json format",
	},
}

var statCmd = &cli.Command{
	Name:      "stat",
	Usage:     "get file or storage info",
	UsageText: "byctl stat [command options] [source]",
	Flags:     mergeFlags(globalFlags, statFlags),
	Before: func(c *cli.Context) error {
		if args := c.Args().Len(); args < 1 {
			return fmt.Errorf("stat command wants one args, but got %d", args)
		}
		return nil
	},
	Action: func(c *cli.Context) (err error) {
		logger, _ := zap.NewDevelopment()

		cfg, err := loadConfig(c, true)
		if err != nil {
			logger.Error("load config", zap.Error(err))
			return err
		}

		conn, key, err := cfg.ParseProfileInput(c.Args().Get(0))
		if err != nil {
			logger.Error("parse profile input from src", zap.Error(err))
			return err
		}

		store, err := services.NewStoragerFromString(conn)
		if err != nil {
			logger.Error("init src storager", zap.Error(err), zap.String("conn string", conn))
			return err
		}

		so := operations.NewSingleOperator(store)

		format := normalFormat
		if c.Bool(statFlagJson) {
			format = jsonFormat
		}

		if key == "" {
			meta := so.StatStorager()
			sm := parseStorager(meta, conn)

			out, err := sm.FormatStorager(format)
			if err != nil {
				logger.Error("format storager", zap.Error(err))
				return err
			}

			fmt.Println(out)
		} else {
			o, err := so.Stat(key)
			if err != nil {
				logger.Error("stat", zap.Error(err))
				return err
			}

			fm, err := parseFileObject(o)
			if err != nil {
				logger.Error("parse file object", zap.Error(err))
				return err
			}

			out, err := fm.FormatFile(format)
			if err != nil {
				logger.Error("format file", zap.Error(err))
				return err
			}

			fmt.Println(out)
		}

		return
	},
}

const (
	normalFormat = iota
	jsonFormat
)

type fileMessage struct {
	ID             string            // file absolute path
	Path           string            // file relative path
	Mode           string            // mode
	LastModified   time.Time         // lastModified
	ContentLength  int64             // ContentLength
	Etag           string            // Etag
	ContentType    string            // ContentType
	SystemMetadata interface{}       // system metadata
	UserMetadata   map[string]string // user metadata
}

func (fm *fileMessage) FormatFile(layout int) (string, error) {
	switch layout {
	case normalFormat:
		return fm.normalFileFormat()
	case jsonFormat:
		return fm.jsonFileFormat()
	default:
		panic("not support format")
	}
}

func (fm *fileMessage) normalFileFormat() (string, error) {
	buf := pool.Get()
	defer buf.Free()

	buf.AppendString(fmt.Sprintf("ID: %s\n", fm.ID))
	buf.AppendString(fmt.Sprintf("Path: %s\n", fm.Path))
	buf.AppendString(fmt.Sprintf("Mode: %s\n", fm.Mode))
	buf.AppendString(fmt.Sprintf("LastModified: %s\n", fm.LastModified))
	buf.AppendString(fmt.Sprintf("ContentLength: %d\n", fm.ContentLength))
	buf.AppendString(fmt.Sprintf("Etag: %s\n", fm.Etag))
	buf.AppendString(fmt.Sprintf("ContentType: %s\n", fm.ContentType))

	buf.AppendString(fmt.Sprintln("\nSystemMetadata:"))
	// Convert system metadata into a map, so that all the data in it can be retrieved.
	sysMeta, err := json.Marshal(fm.SystemMetadata)
	if err != nil {
		return "", err
	}
	var m map[string]interface{}
	err = json.Unmarshal(sysMeta, &m)
	if err != nil {
		return "", err
	}
	for k, v := range m {
		buf.AppendString(fmt.Sprintf("%s: \"%s\"\n", k, v))
	}

	buf.AppendString(fmt.Sprint("\nUserMetadata: "))
	if fm.UserMetadata == nil {
		buf.AppendString("null\n")
		return buf.String(), nil
	}
	for k, v := range fm.UserMetadata {
		buf.AppendString(fmt.Sprintf("%s: \"%s\"\n", k, v))
	}

	return buf.String(), nil
}

func (fm *fileMessage) jsonFileFormat() (string, error) {
	b, err := json.Marshal(&fm)
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "    ")
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

type storageMessage struct {
	Service  string // service name
	Name     string // bucket name
	WorkDir  string // work dir
	Location string // bucket location
}

func (sm *storageMessage) FormatStorager(layout int) (string, error) {
	switch layout {
	case normalFormat:
		return sm.normalStoragerFormat()
	case jsonFormat:
		return sm.jsonStoragerFormat()
	default:
		panic("not support format")
	}
}

func (sm *storageMessage) normalStoragerFormat() (string, error) {
	buf := pool.Get()
	defer buf.Free()

	buf.AppendString(fmt.Sprintf("Service: %s\n", sm.Service))
	buf.AppendString(fmt.Sprintf("Name: %s\n", sm.Name))
	buf.AppendString(fmt.Sprintf("WorkDir: %s", sm.WorkDir))

	if sm.Location != "" {
		buf.AppendString(fmt.Sprintf("\nLocation: %s", sm.Location))
	}

	return buf.String(), nil
}

func (sm *storageMessage) jsonStoragerFormat() (string, error) {
	b, err := json.Marshal(sm)
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "    ")
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func parseFileObject(o *types.Object) (*fileMessage, error) {
	fm := &fileMessage{
		ID:   o.ID,
		Path: o.Path,
	}

	fm.Mode = o.Mode.String()

	if v, ok := o.GetLastModified(); ok {
		fm.LastModified = v
	}
	if v, ok := o.GetContentLength(); ok {
		fm.ContentLength = v
	}
	if v, ok := o.GetEtag(); ok {
		fm.Etag = v
	}
	if v, ok := o.GetContentType(); ok {
		fm.ContentType = v
	}
	if v, ok := o.GetSystemMetadata(); ok {
		fm.SystemMetadata = v
	}
	if v, ok := o.GetUserMetadata(); ok {
		fm.UserMetadata = v
	}

	return fm, nil
}

func parseStorager(meta *types.StorageMeta, conn string) *storageMessage {
	sm := &storageMessage{
		Name:    meta.Name,
		WorkDir: meta.WorkDir,
	}

	if v, ok := meta.GetLocation(); ok {
		sm.Location = v
	}

	// Get service name by conn.
	// For example: conn = s3://bucketname/workdir?credential=xxx&endpoint=xxx&location=xxx
	// We can get "s3".
	index := strings.Index(conn, ":")
	serviceName := conn[:index]
	sm.Service = serviceName

	return sm
}
