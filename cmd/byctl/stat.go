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
	ID             string                 `json:"id"`             // file absolute path
	Path           string                 `json:"path"`           // file relative path
	Mode           string                 `json:"mode"`           // mode
	LastModified   time.Time              `json:"lastModified"`   // lastModified
	ContentLength  int64                  `json:"contentLength"`  // ContentLength
	Etag           string                 `json:"etag"`           // Etag
	ContentType    string                 `json:"contentType"`    // ContentType
	SystemMetadata map[string]interface{} `json:"systemMetadata"` // system metadata
	UserMetadata   map[string]string      `json:"userMetadata"`   // user metadata
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
	for k, v := range fm.SystemMetadata {
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
	Service    string `json:"service"`    // service name
	BucketName string `json:"bucketName"` // bucket name
	WorkDir    string `json:"workDir"`    // work dir
	Location   string `json:"location"`   // bucket location
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
	buf.AppendString(fmt.Sprintf("BucketName: %s\n", sm.BucketName))
	buf.AppendString(fmt.Sprintf("WorkDir: %s\n", sm.WorkDir))

	if sm.Location != "" {
		buf.AppendString(fmt.Sprintf("Location: %s", sm.Location))
	} else {
		buf.AppendString(fmt.Sprintf("Location: null"))
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

	var modes string

	// An object may have more than one mode.
	if o.Mode.IsDir() {
		modes += "Dir|"
	}
	if o.Mode.IsRead() {
		modes += "Read|"
	}
	if o.Mode.IsLink() {
		modes += "Link|"
	}
	if o.Mode.IsAppend() {
		modes += "Append|"
	}
	if o.Mode.IsBlock() {
		modes += "Block|"
	}
	if o.Mode.IsPage() {
		modes += "Page"
	}
	if o.Mode.IsPart() {
		modes += "Part|"
	}

	// Remove the `|` at the end of modes.
	mode := strings.TrimSuffix(modes, "|")
	fm.Mode = mode

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
	// Convert system metadata into a map, so that all the data in it can be retrieved.
	if v, ok := o.GetSystemMetadata(); ok {
		sysMeta, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		var m map[string]interface{}

		err = json.Unmarshal(sysMeta, &m)
		if err != nil {
			return nil, err
		}

		fm.SystemMetadata = m
	}
	if v, ok := o.GetUserMetadata(); ok {
		fm.UserMetadata = v
	}

	return fm, nil
}

func parseStorager(meta *types.StorageMeta, conn string) *storageMessage {
	sm := &storageMessage{
		BucketName: meta.Name,
		WorkDir:    meta.WorkDir,
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
