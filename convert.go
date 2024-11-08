package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	"github.com/itimofeev/go-saga"
)

var mdImageRe = regexp.MustCompile(`!\[([^\[\]]*?)\]\((.*?)\)`)
var urlGroup = 2

func convert(ctx context.Context, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("can't read file %s: %w", filePath, err)
	}

	matches := mdImageRe.FindAllSubmatchIndex(content, -1)
	if matches == nil {
		// if images are not found then do nothing
		return nil
	}

	// Convert markdown and collect paths

	workingDirectory := filepath.Dir(filePath)
	renameMapping := make(map[string]string, 0)

	var result bytes.Buffer
	var leftCursor int = 0
	var rightCursor int

	for _, match := range matches {
		// write unchanged content before path
		rightCursor = match[urlGroup*2]

		_, err = result.Write(content[leftCursor:rightCursor])
		if err != nil {
			return fmt.Errorf("can't write result buffer: %w", err)
		}

		// update path
		leftCursor = rightCursor
		rightCursor = match[urlGroup*2+1]

		u, err := url.Parse(string(content[leftCursor:rightCursor]))
		if err != nil {
			return fmt.Errorf("can't unescape path %s: %w", content[leftCursor:rightCursor], err)
		}

		absPath := filepath.Join(workingDirectory, u.Path)

		if stat, err := os.Stat(absPath); err == nil && !stat.IsDir() { // path is for file and file is accessible
			newFileName := transliterateRussian(filepath.Base(u.Path))

			u.Path = filepath.Join(filepath.Dir(u.Path), newFileName)
			renameMapping[absPath] = filepath.Join(filepath.Dir(absPath), newFileName)

			_, err = result.Write([]byte(u.String()))
			if err != nil {
				return fmt.Errorf("can't write result buffer: %w", err)
			}
		} else { // else write with no changes
			_, err = result.Write(content[leftCursor:rightCursor])
			if err != nil {
				return fmt.Errorf("can't write result buffer: %w", err)
			}
		}

		// move cursor after path
		leftCursor = rightCursor
	}

	_, err = result.Write(content[leftCursor:])
	if err != nil {
		return fmt.Errorf("can't write result buffer: %w", err)
	}

	s := saga.NewSaga("convert")

	for oldPath, newPath := range renameMapping {
		err = s.AddStep(&saga.Step{
			Name: fmt.Sprintf("rename %s -> %s", oldPath, newPath),
			Func: func(context.Context) error {
				return os.Rename(oldPath, newPath)
			},
			CompensateFunc: func(context.Context) error {
				return os.Rename(newPath, oldPath)
			},
		})
		if err != nil {
			return fmt.Errorf("can't add rename %s -> %s step to saga: %w", oldPath, newPath, err)
		}
	}

	err = s.AddStep(&saga.Step{
		Name: "update markdown",
		Func: func(context.Context) error {
			return os.WriteFile(filePath, result.Bytes(), 0644)
		},
		CompensateFunc: func(context.Context) error {
			return os.WriteFile(filePath, content, 0644)
		},
	})
	if err != nil {
		return fmt.Errorf("can't add update markdown step to saga: %w", err)
	}

	store := saga.New()
	c := saga.NewCoordinator(ctx, ctx, s, store)

	sr := c.Play()
	if sr.ExecutionError != nil {
		return fmt.Errorf("can't convert: %w (compensate errors: %w)", sr.ExecutionError, errors.Join(sr.CompensateErrors...))
	}

	return nil
}
