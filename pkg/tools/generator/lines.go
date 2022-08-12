package generator

import (
	"bufio"
	"context"
	"github.com/pkg/errors"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func WriteLines(ctx context.Context, w io.Writer, lineCount int, lineBytes int) error {
	writer := bufio.NewWriter(w)

	// TODO length of line bytes would be larger then r.config.lineBytes, cause we ignore the time prefix
	line := make([]byte, lineBytes)
	for i := 0; i < len(line); i++ {
		line[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	lineStr := string(line)

	const flushMax = 100
	flushCount := 0
	builder := strings.Builder{}
	for i := 1; i <= lineCount; i++ {
		builder.Reset()

		builder.WriteString(strconv.Itoa(i))
		builder.WriteString(" ")
		builder.WriteString(time.Now().Format(time.RFC3339))
		builder.WriteString(" ")
		builder.WriteString(lineStr)
		builder.WriteString("\n")

		if _, err := writer.WriteString(builder.String()); err != nil {
			return err
		}

		flushCount++
		if flushCount == flushMax {
			select {
			case <-ctx.Done():
				return errors.New("write file timeout")
			default:
			}

			if err := writer.Flush(); err != nil {
				return errors.WithMessage(err, "flush failed")
			}

			flushCount = 0
		}
	}
	if err := writer.Flush(); err != nil {
		return errors.WithMessage(err, "flush failed")
	}

	return nil
}
