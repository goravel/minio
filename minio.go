package minio

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gookit/color"
	"github.com/goravel/framework/support/carbon"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/support/str"
)

/*
 * MinIO OSS
 * Document: https://min.io/docs/minio/linux/developers/go/minio-go.html
 * Example: https://github.com/minio/minio-go/tree/master/examples/s3
 */

type Minio struct {
	ctx      context.Context
	config   config.Config
	instance *minio.Client
	bucket   string
	disk     string
	url      string
}

func NewMinio(ctx context.Context, config config.Config, disk string) (*Minio, error) {
	key := config.GetString(fmt.Sprintf("filesystems.disks.%s.key", disk))
	secret := config.GetString(fmt.Sprintf("filesystems.disks.%s.secret", disk))
	region := config.GetString(fmt.Sprintf("filesystems.disks.%s.region", disk))
	bucket := config.GetString(fmt.Sprintf("filesystems.disks.%s.bucket", disk))
	diskUrl := config.GetString(fmt.Sprintf("filesystems.disks.%s.url", disk))
	ssl := config.GetBool(fmt.Sprintf("filesystems.disks.%s.ssl", disk), false)
	endpoint := config.GetString(fmt.Sprintf("filesystems.disks.%s.endpoint", disk))
	if key == "" || secret == "" || bucket == "" || diskUrl == "" || endpoint == "" {
		return nil, fmt.Errorf("please set %s configuration first", disk)
	}

	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(key, secret, ""),
		Secure: ssl,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("init %s disk error: %s", disk, err)
	}

	return &Minio{
		ctx:      ctx,
		config:   config,
		instance: client,
		bucket:   bucket,
		disk:     disk,
		url:      diskUrl,
	}, nil
}

func (r *Minio) AllDirectories(path string) ([]string, error) {
	var directories []string
	validPath := validPath(path)
	objectCh := r.instance.ListObjects(r.ctx, r.bucket, minio.ListObjectsOptions{
		Prefix:    validPath,
		Recursive: false,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}

		if strings.HasSuffix(object.Key, "/") {
			key := strings.TrimPrefix(object.Key, validPath)
			if key != "" {
				directories = append(directories, key)
				subDirectories, err := r.AllDirectories(object.Key)
				if err != nil {
					return nil, err
				}
				for _, subDirectory := range subDirectories {
					directories = append(directories, strings.TrimPrefix(object.Key+subDirectory, validPath))
				}
			}
		}
	}

	return directories, nil
}

func (r *Minio) AllFiles(path string) ([]string, error) {
	var files []string
	validPath := validPath(path)

	objectCh := r.instance.ListObjects(r.ctx, r.bucket, minio.ListObjectsOptions{
		Prefix:    validPath,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}

		if !strings.HasSuffix(object.Key, "/") {
			files = append(files, strings.TrimPrefix(object.Key, validPath))
		}
	}

	return files, nil
}

func (r *Minio) Copy(originFile, targetFile string) error {
	srcOpts := minio.CopySrcOptions{
		Bucket: r.bucket,
		Object: originFile,
	}
	dstOpts := minio.CopyDestOptions{
		Bucket: r.bucket,
		Object: targetFile,
	}
	_, err := r.instance.CopyObject(r.ctx, dstOpts, srcOpts)
	return err
}

func (r *Minio) Delete(files ...string) error {
	objectsCh := make(chan minio.ObjectInfo, len(files))
	go func() {
		defer close(objectsCh)
		for _, file := range files {
			object := minio.ObjectInfo{
				Key: file,
			}
			objectsCh <- object
		}
	}()

	for err := range r.instance.RemoveObjects(r.ctx, r.bucket, objectsCh, minio.RemoveObjectsOptions{}) {
		return err.Err
	}

	return nil
}

func (r *Minio) DeleteDirectory(directory string) error {
	if !strings.HasSuffix(directory, "/") {
		directory += "/"
	}
	opts := minio.RemoveObjectOptions{
		ForceDelete: true,
	}
	err := r.instance.RemoveObject(r.ctx, r.bucket, directory, opts)
	if err != nil {
		return err
	}

	return nil
}

func (r *Minio) Directories(path string) ([]string, error) {
	var directories []string
	validPath := validPath(path)
	objectCh := r.instance.ListObjects(r.ctx, r.bucket, minio.ListObjectsOptions{
		Prefix:    validPath,
		Recursive: false,
	})
	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}
		if strings.HasSuffix(object.Key, "/") {
			directory := strings.ReplaceAll(object.Key, validPath, "")
			if directory != "" {
				directories = append(directories, directory)
			}
		}
	}

	return directories, nil
}

func (r *Minio) Exists(file string) bool {
	_, err := r.instance.StatObject(r.ctx, r.bucket, file, minio.StatObjectOptions{})

	return err == nil
}

func (r *Minio) Files(path string) ([]string, error) {
	var files []string
	validPath := validPath(path)

	for object := range r.instance.ListObjects(r.ctx, r.bucket, minio.ListObjectsOptions{
		Prefix:    validPath,
		Recursive: false,
	}) {
		if object.Err != nil {
			return nil, object.Err
		}
		if !strings.HasSuffix(object.Key, "/") {
			files = append(files, strings.ReplaceAll(object.Key, validPath, ""))
		}
	}

	return files, nil
}

func (r *Minio) Get(file string) (string, error) {
	data, err := r.GetBytes(file)

	return string(data), err
}

func (r *Minio) GetBytes(file string) ([]byte, error) {
	object, err := r.instance.GetObject(r.ctx, r.bucket, file, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(object)
	defer object.Close()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Minio) LastModified(file string) (time.Time, error) {
	objInfo, err := r.instance.StatObject(r.ctx, r.bucket, file, minio.StatObjectOptions{})
	if err != nil {
		return time.Time{}, err
	}

	l, err := time.LoadLocation(r.config.GetString("app.timezone"))
	if err != nil {
		return time.Time{}, err
	}

	return objInfo.LastModified.In(l), nil
}

func (r *Minio) MakeDirectory(directory string) error {
	if !strings.HasSuffix(directory, "/") {
		directory += "/"
	}

	return r.Put(directory, "")
}

func (r *Minio) MimeType(file string) (string, error) {
	objInfo, err := r.instance.StatObject(r.ctx, r.bucket, file, minio.StatObjectOptions{})
	if err != nil {
		return "", err
	}

	return objInfo.ContentType, nil
}

func (r *Minio) Missing(file string) bool {
	return !r.Exists(file)
}

func (r *Minio) Move(oldFile, newFile string) error {
	if err := r.Copy(oldFile, newFile); err != nil {
		return err
	}

	return r.Delete(oldFile)
}

func (r *Minio) Path(file string) string {
	return file
}

func (r *Minio) Put(file string, content string) error {
	// If the file is created in a folder directly, we can't check if the folder exists.
	// So we need to create the top folder first.
	if !strings.HasSuffix(file, "/") {
		index := strings.Index(file, "/")
		if index != -1 {
			folder := file[:index+1]
			if err := r.MakeDirectory(folder); err != nil {
				return err
			}
		}
	}

	mtype := mimetype.Detect([]byte(content))
	reader := strings.NewReader(content)
	_, err := r.instance.PutObject(
		r.ctx,
		r.bucket,
		file,
		reader,
		reader.Size(),
		minio.PutObjectOptions{
			ContentType: mtype.String(),
		},
	)

	return err
}

func (r *Minio) PutFile(filePath string, source filesystem.File) (string, error) {
	return r.PutFileAs(filePath, source, str.Random(40))
}

func (r *Minio) PutFileAs(filePath string, source filesystem.File, name string) (string, error) {
	fullPath, err := fullPathOfFile(filePath, source, name)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(source.File())
	if err != nil {
		return "", err
	}

	if err := r.Put(fullPath, string(data)); err != nil {
		return "", err
	}

	return fullPath, nil
}

func (r *Minio) Size(file string) (int64, error) {
	objInfo, err := r.instance.StatObject(r.ctx, r.bucket, file, minio.StatObjectOptions{})
	if err != nil {
		return 0, err
	}

	return objInfo.Size, nil
}

func (r *Minio) TemporaryUrl(file string, time time.Time) (string, error) {
	file = strings.TrimPrefix(file, "/")
	reqParams := make(url.Values)
	presignedURL, err := r.instance.PresignedGetObject(r.ctx, r.bucket, file, time.Sub(carbon.Now().StdTime()), reqParams)
	if err != nil {
		return "", err
	}

	return presignedURL.String(), nil
}

func (r *Minio) WithContext(ctx context.Context) filesystem.Driver {
	driver, err := NewMinio(ctx, r.config, r.disk)
	if err != nil {
		color.Redf("init %s disk fail: %v\n", r.disk, err)

		return nil
	}

	return driver
}

func (r *Minio) Url(file string) string {
	realUrl := strings.TrimSuffix(r.url, "/")
	if !strings.HasSuffix(realUrl, r.bucket) {
		realUrl += "/" + r.bucket
	}

	return realUrl + "/" + strings.TrimPrefix(file, "/")
}
