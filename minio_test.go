package minio

import (
	"context"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gookit/color"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
)

func TestStorage(t *testing.T) {
	if os.Getenv("MINIO_ACCESS_KEY_ID") == "" {
		color.Redln("No filesystem tests run, please add minio configuration: MINIO_ACCESS_KEY_ID= MINIO_ACCESS_KEY_SECRET= MINIO_BUCKET= go test ./...")
		return
	}

	assert.Nil(t, ioutil.WriteFile("test.txt", []byte("Goravel"), 0644))

	mockConfig := &configmocks.Config{}
	mockConfig.On("GetString", "app.timezone").Return("UTC")
	mockConfig.On("GetString", "minio.driver").Return("minio")
	mockConfig.On("GetString", "minio.region").Return("")
	mockConfig.On("GetBool", "minio.ssl", false).Return(false)
	mockConfig.On("GetString", "minio.key").Return(os.Getenv("MINIO_ACCESS_KEY_ID"))
	mockConfig.On("GetString", "minio.secret").Return(os.Getenv("MINIO_ACCESS_KEY_SECRET"))
	mockConfig.On("GetString", "minio.bucket").Return(os.Getenv("MINIO_BUCKET"))

	minioPool, minioResource, err := initMinioDocker(mockConfig)

	var driver contractsfilesystem.Driver
	url := mockConfig.GetString("minio.url")

	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "AllDirectories",
			setup: func() {
				assert.Nil(t, driver.Put("AllDirectories/1.txt", "Goravel"))
				assert.Nil(t, driver.Put("AllDirectories/2.txt", "Goravel"))
				assert.Nil(t, driver.Put("AllDirectories/3/3.txt", "Goravel"))
				assert.Nil(t, driver.Put("AllDirectories/3/5/6/6.txt", "Goravel"))
				assert.Nil(t, driver.MakeDirectory("AllDirectories/3/4"))
				assert.True(t, driver.Exists("AllDirectories/1.txt"))
				assert.True(t, driver.Exists("AllDirectories/2.txt"))
				assert.True(t, driver.Exists("AllDirectories/3/3.txt"))
				assert.True(t, driver.Exists("AllDirectories/3/4/"))
				assert.True(t, driver.Exists("AllDirectories/3/5/6/6.txt"))
				files, err := driver.AllDirectories("AllDirectories")
				assert.Nil(t, err)
				assert.Equal(t, []string{"3/", "3/4/", "3/5/", "3/5/6/"}, files)
				files, err = driver.AllDirectories("./AllDirectories")
				assert.Nil(t, err)
				assert.Equal(t, []string{"3/", "3/4/", "3/5/", "3/5/6/"}, files)
				files, err = driver.AllDirectories("/AllDirectories")
				assert.Nil(t, err)
				assert.Equal(t, []string{"3/", "3/4/", "3/5/", "3/5/6/"}, files)
				files, err = driver.AllDirectories("./AllDirectories/")
				assert.Nil(t, err)
				assert.Equal(t, []string{"3/", "3/4/", "3/5/", "3/5/6/"}, files)
				assert.Nil(t, driver.DeleteDirectory("AllDirectories"))
			},
		},
		{
			name: "AllFiles",
			setup: func() {
				assert.Nil(t, driver.Put("AllFiles/1.txt", "Goravel"))
				assert.Nil(t, driver.Put("AllFiles/2.txt", "Goravel"))
				assert.Nil(t, driver.Put("AllFiles/3/3.txt", "Goravel"))
				assert.Nil(t, driver.Put("AllFiles/3/4/4.txt", "Goravel"))
				assert.True(t, driver.Exists("AllFiles/1.txt"))
				assert.True(t, driver.Exists("AllFiles/2.txt"))
				assert.True(t, driver.Exists("AllFiles/3/3.txt"))
				assert.True(t, driver.Exists("AllFiles/3/4/4.txt"))
				files, err := driver.AllFiles("AllFiles")
				assert.Nil(t, err)
				assert.Equal(t, []string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files)
				files, err = driver.AllFiles("./AllFiles")
				assert.Nil(t, err)
				assert.Equal(t, []string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files)
				files, err = driver.AllFiles("/AllFiles")
				assert.Nil(t, err)
				assert.Equal(t, []string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files)
				files, err = driver.AllFiles("./AllFiles/")
				assert.Nil(t, err)
				assert.Equal(t, []string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files)
				assert.Nil(t, driver.DeleteDirectory("AllFiles"))
			},
		},
		{
			name: "Copy",
			setup: func() {
				assert.Nil(t, driver.Put("Copy/1.txt", "Goravel"))
				assert.True(t, driver.Exists("Copy/1.txt"))
				assert.Nil(t, driver.Copy("Copy/1.txt", "Copy1/1.txt"))
				assert.True(t, driver.Exists("Copy/1.txt"))
				assert.True(t, driver.Exists("Copy1/1.txt"))
				assert.Nil(t, driver.DeleteDirectory("Copy"))
				assert.Nil(t, driver.DeleteDirectory("Copy1"))
			},
		},
		{
			name: "Delete",
			setup: func() {
				assert.Nil(t, driver.Put("Delete/1.txt", "Goravel"))
				assert.True(t, driver.Exists("Delete/1.txt"))
				assert.Nil(t, driver.Delete("Delete/1.txt"))
				assert.True(t, driver.Missing("Delete/1.txt"))
				assert.Nil(t, driver.DeleteDirectory("Delete"))
			},
		},
		{
			name: "DeleteDirectory",
			setup: func() {
				assert.Nil(t, driver.Put("DeleteDirectory/1.txt", "Goravel"))
				assert.True(t, driver.Exists("DeleteDirectory/1.txt"))
				assert.Nil(t, driver.DeleteDirectory("DeleteDirectory"))
				assert.True(t, driver.Missing("DeleteDirectory/1.txt"))
				assert.Nil(t, driver.DeleteDirectory("DeleteDirectory"))
			},
		},
		{
			name: "Directories",
			setup: func() {
				assert.Nil(t, driver.Put("Directories/1.txt", "Goravel"))
				assert.Nil(t, driver.Put("Directories/2.txt", "Goravel"))
				assert.Nil(t, driver.Put("Directories/3/3.txt", "Goravel"))
				assert.Nil(t, driver.Put("Directories/3/5/5.txt", "Goravel"))
				assert.Nil(t, driver.MakeDirectory("Directories/3/4"))
				assert.True(t, driver.Exists("Directories/1.txt"))
				assert.True(t, driver.Exists("Directories/2.txt"))
				assert.True(t, driver.Exists("Directories/3/3.txt"))
				assert.True(t, driver.Exists("Directories/3/4/"))
				assert.True(t, driver.Exists("Directories/3/5/5.txt"))
				files, err := driver.Directories("Directories")
				assert.Nil(t, err)
				assert.Equal(t, []string{"3/"}, files)
				files, err = driver.Directories("./Directories")
				assert.Nil(t, err)
				assert.Equal(t, []string{"3/"}, files)
				files, err = driver.Directories("/Directories")
				assert.Nil(t, err)
				assert.Equal(t, []string{"3/"}, files)
				files, err = driver.Directories("./Directories/")
				assert.Nil(t, err)
				assert.Equal(t, []string{"3/"}, files)
				assert.Nil(t, driver.DeleteDirectory("Directories"))
			},
		},
		{
			name: "Files",
			setup: func() {
				assert.Nil(t, driver.Put("Files/1.txt", "Goravel"))
				assert.Nil(t, driver.Put("Files/2.txt", "Goravel"))
				assert.Nil(t, driver.Put("Files/3/3.txt", "Goravel"))
				assert.Nil(t, driver.Put("Files/3/4/4.txt", "Goravel"))
				assert.True(t, driver.Exists("Files/1.txt"))
				assert.True(t, driver.Exists("Files/2.txt"))
				assert.True(t, driver.Exists("Files/3/3.txt"))
				assert.True(t, driver.Exists("Files/3/4/4.txt"))
				files, err := driver.Files("Files")
				assert.Nil(t, err)
				assert.Equal(t, []string{"1.txt", "2.txt"}, files)
				files, err = driver.Files("./Files")
				assert.Nil(t, err)
				assert.Equal(t, []string{"1.txt", "2.txt"}, files)
				files, err = driver.Files("/Files")
				assert.Nil(t, err)
				assert.Equal(t, []string{"1.txt", "2.txt"}, files)
				files, err = driver.Files("./Files/")
				assert.Nil(t, err)
				assert.Equal(t, []string{"1.txt", "2.txt"}, files)
				assert.Nil(t, driver.DeleteDirectory("Files"))
			},
		},
		{
			name: "Get",
			setup: func() {
				assert.Nil(t, driver.Put("Get/1.txt", "Goravel"))
				assert.True(t, driver.Exists("Get/1.txt"))
				data, err := driver.Get("Get/1.txt")
				assert.Nil(t, err)
				assert.Equal(t, "Goravel", data)
				length, err := driver.Size("Get/1.txt")
				assert.Nil(t, err)
				assert.Equal(t, int64(7), length)
				assert.Nil(t, driver.DeleteDirectory("Get"))
			},
		},
		{
			name: "LastModified",
			setup: func() {
				assert.Nil(t, driver.Put("LastModified/1.txt", "Goravel"))
				assert.True(t, driver.Exists("LastModified/1.txt"))
				date, err := driver.LastModified("LastModified/1.txt")
				assert.Nil(t, err)

				l, err := time.LoadLocation("UTC")
				assert.Nil(t, err)
				assert.Equal(t, time.Now().In(l).Format("2006-01-02 15"), date.Format("2006-01-02 15"))
				assert.Nil(t, driver.DeleteDirectory("LastModified"))
			},
		},
		{
			name: "MakeDirectory",
			setup: func() {
				assert.Nil(t, driver.MakeDirectory("MakeDirectory1/"))
				assert.Nil(t, driver.MakeDirectory("MakeDirectory2"))
				assert.Nil(t, driver.MakeDirectory("MakeDirectory3/MakeDirectory4"))
				assert.Nil(t, driver.DeleteDirectory("MakeDirectory1"))
				assert.Nil(t, driver.DeleteDirectory("MakeDirectory2"))
				assert.Nil(t, driver.DeleteDirectory("MakeDirectory3"))
				assert.Nil(t, driver.DeleteDirectory("MakeDirectory4"))
			},
		},
		{
			name: "MimeType",
			setup: func() {
				assert.Nil(t, driver.Put("MimeType/1.txt", "Goravel"))
				assert.True(t, driver.Exists("MimeType/1.txt"))
				mimeType, err := driver.MimeType("MimeType/1.txt")
				assert.Nil(t, err)
				mediaType, _, err := mime.ParseMediaType(mimeType)
				assert.Nil(t, err)
				assert.Equal(t, "text/plain", mediaType)

				fileInfo := &File{path: "logo.png"}
				path, err := driver.PutFile("MimeType", fileInfo)
				assert.Nil(t, err)
				assert.True(t, driver.Exists(path))
				mimeType, err = driver.MimeType(path)
				assert.Nil(t, err)
				assert.Equal(t, "image/png", mimeType)
			},
		},
		{
			name: "Move",
			setup: func() {
				assert.Nil(t, driver.Put("Move/1.txt", "Goravel"))
				assert.True(t, driver.Exists("Move/1.txt"))
				assert.Nil(t, driver.Move("Move/1.txt", "Move1/1.txt"))
				assert.True(t, driver.Missing("Move/1.txt"))
				assert.True(t, driver.Exists("Move1/1.txt"))
				assert.Nil(t, driver.DeleteDirectory("Move"))
				assert.Nil(t, driver.DeleteDirectory("Move1"))
			},
		},
		{
			name: "Put",
			setup: func() {
				assert.Nil(t, driver.Put("Put/1.txt", "Goravel"))
				assert.True(t, driver.Exists("Put/1.txt"))
				assert.True(t, driver.Missing("Put/2.txt"))
				assert.Nil(t, driver.DeleteDirectory("Put"))
			},
		},
		{
			name: "PutFile_Image",
			setup: func() {
				fileInfo := &File{path: "logo.png"}
				path, err := driver.PutFile("PutFile1", fileInfo)
				assert.Nil(t, err)
				assert.True(t, driver.Exists(path))
				assert.Nil(t, driver.DeleteDirectory("PutFile1"))
			},
		},
		{
			name: "PutFile_Text",
			setup: func() {
				fileInfo := &File{path: "test.txt"}
				path, err := driver.PutFile("PutFile", fileInfo)
				assert.Nil(t, err)
				assert.True(t, driver.Exists(path))
				data, err := driver.Get(path)
				assert.Nil(t, err)
				assert.Equal(t, "Goravel", data)
				assert.Nil(t, driver.DeleteDirectory("PutFile"))
			},
		},
		{
			name: "PutFileAs_Text",
			setup: func() {
				fileInfo := &File{path: "test.txt"}
				path, err := driver.PutFileAs("PutFileAs", fileInfo, "text")
				assert.Nil(t, err)
				assert.Equal(t, "PutFileAs/text.txt", path)
				assert.True(t, driver.Exists(path))
				data, err := driver.Get(path)
				assert.Nil(t, err)
				assert.Equal(t, "Goravel", data)

				path, err = driver.PutFileAs("PutFileAs", fileInfo, "text1.txt")
				assert.Nil(t, err)
				assert.Equal(t, "PutFileAs/text1.txt", path)
				assert.True(t, driver.Exists(path))
				data, err = driver.Get(path)
				assert.Nil(t, err)
				assert.Equal(t, "Goravel", data)

				assert.Nil(t, driver.DeleteDirectory("PutFileAs"))
			},
		},
		{
			name: "PutFileAs_Image",
			setup: func() {
				fileInfo := &File{path: "logo.png"}
				path, err := driver.PutFileAs("PutFileAs1", fileInfo, "image")
				assert.Nil(t, err)
				assert.Equal(t, "PutFileAs1/image.png", path)
				assert.True(t, driver.Exists(path))

				path, err = driver.PutFileAs("PutFileAs1", fileInfo, "image1.png")
				assert.Nil(t, err)
				assert.Equal(t, "PutFileAs1/image1.png", path)
				assert.True(t, driver.Exists(path))

				assert.Nil(t, driver.DeleteDirectory("PutFileAs1"))
			},
		},
		{
			name: "Size",
			setup: func() {
				assert.Nil(t, driver.Put("Size/1.txt", "Goravel"))
				assert.True(t, driver.Exists("Size/1.txt"))
				length, err := driver.Size("Size/1.txt")
				assert.Nil(t, err)
				assert.Equal(t, int64(7), length)
				assert.Nil(t, driver.DeleteDirectory("Size"))
			},
		},
		{
			name: "TemporaryUrl",
			setup: func() {
				assert.Nil(t, driver.Put("TemporaryUrl/1.txt", "Goravel"))
				assert.True(t, driver.Exists("TemporaryUrl/1.txt"))
				url, err := driver.TemporaryUrl("TemporaryUrl/1.txt", time.Now().Add(5*time.Second))
				assert.Nil(t, err)
				assert.NotEmpty(t, url)
				resp, err := http.Get(url)
				assert.Nil(t, err)
				content, err := ioutil.ReadAll(resp.Body)
				assert.Nil(t, resp.Body.Close())
				assert.Nil(t, err)
				assert.Equal(t, "Goravel", string(content))
				assert.Nil(t, driver.DeleteDirectory("TemporaryUrl"))
			},
		},
		{
			name: "Url",
			setup: func() {
				assert.Nil(t, driver.Put("Url/1.txt", "Goravel"))
				assert.True(t, driver.Exists("Url/1.txt"))
				url := url + "/Url/1.txt"
				assert.Equal(t, url, driver.Url("Url/1.txt"))
				resp, err := http.Get(url)
				assert.Nil(t, err)
				content, err := ioutil.ReadAll(resp.Body)
				assert.Nil(t, resp.Body.Close())
				assert.Nil(t, err)
				assert.Equal(t, "Goravel", string(content))
				assert.Nil(t, driver.DeleteDirectory("Url"))
			},
		},
	}

	driver, err = NewMinio(context.Background(), mockConfig)
	assert.NotNil(t, driver)
	assert.Nil(t, err)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()
		})
	}

	assert.Nil(t, os.Remove("test.txt"))
	assert.Nil(t, minioPool.Purge(minioResource))
}

type File struct {
	path string
}

func (f *File) Disk(disk string) contractsfilesystem.File {
	return &File{}
}

func (f *File) Extension() (string, error) {
	return "", nil
}

func (f *File) File() string {
	return f.path
}

func (f *File) GetClientOriginalName() string {
	return ""
}

func (f *File) GetClientOriginalExtension() string {
	return ""
}

func (f *File) HashName(path ...string) string {
	return ""
}

func (f *File) LastModified() (time.Time, error) {
	return time.Now(), nil
}

func (f *File) MimeType() (string, error) {
	return "", nil
}

func (f *File) Size() (int64, error) {
	return 0, nil
}

func (f *File) Store(path string) (string, error) {
	return "", nil
}

func (f *File) StoreAs(path string, name string) (string, error) {
	return "", nil
}

func pool() (*dockertest.Pool, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, errors.WithMessage(err, "Could not construct pool")
	}

	if err := pool.Client.Ping(); err != nil {
		return nil, errors.WithMessage(err, "Could not connect to Docker")
	}

	return pool, nil
}

func resource(pool *dockertest.Pool, opts *dockertest.RunOptions) (*dockertest.Resource, error) {
	return pool.RunWithOptions(opts, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
}

func initMinioDocker(mockConfig *configmocks.Config) (*dockertest.Pool, *dockertest.Resource, error) {
	pool, err := pool()
	if err != nil {
		return nil, nil, err
	}

	key := mockConfig.GetString("minio.key")
	secret := mockConfig.GetString("minio.secret")
	bucket := mockConfig.GetString("minio.bucket")
	resource, err := resource(pool, &dockertest.RunOptions{
		Repository: "minio/minio",
		Tag:        "latest",
		Env: []string{
			"MINIO_ACCESS_KEY=" + key,
			"MINIO_SECRET_KEY=" + secret,
		},
		Cmd: []string{
			"server",
			"/data",
		},
		ExposedPorts: []string{
			"9000/tcp",
		},
	})
	if err != nil {
		return nil, nil, err
	}

	_ = resource.Expire(600)

	endpoint := fmt.Sprintf("127.0.0.1:%s", resource.GetPort("9000/tcp"))
	mockConfig.On("GetString", "minio.url").Return(fmt.Sprintf("http://%s/%s", endpoint, bucket))
	mockConfig.On("GetString", "minio.endpoint").Return(endpoint)

	if err := pool.Retry(func() error {
		client, err := minio.New(endpoint, &minio.Options{
			Creds: credentials.NewStaticV4(key, secret, ""),
		})
		if err != nil {
			return err
		}

		if err := client.MakeBucket(context.Background(), mockConfig.GetString("minio.bucket"), minio.MakeBucketOptions{}); err != nil {
			return err
		}

		policy := `{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Action": [
                    "s3:GetObject",
                    "s3:PutObject"
                ],
                "Effect": "Allow",
                "Principal": "*",
                "Resource": [
                    "arn:aws:s3:::` + bucket + `/*"
                ]
            },
            {
                "Action": [
                    "s3:ListBucket"
                ],
                "Effect": "Allow",
                "Principal": "*",
                "Resource": [
                    "arn:aws:s3:::` + bucket + `"
                ]
            }
        ]
    }`

		if err := client.SetBucketPolicy(context.Background(), bucket, policy); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	return pool, resource, nil
}
