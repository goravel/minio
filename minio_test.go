package minio

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"testing"
	"time"

	filesystemcontract "github.com/goravel/framework/contracts/filesystem"
	contractsdocker "github.com/goravel/framework/contracts/testing/docker"
	configmock "github.com/goravel/framework/mocks/config"
	supportdocker "github.com/goravel/framework/support/docker"
	testingdocker "github.com/goravel/framework/testing/docker"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/suite"
)

const (
	testKey    = "j45vy0yIvvhs47uf"
	testSecret = "jgwB91UEJ1LBA9cVYfbVqP7eXpHHrpDs"
	testBucket = "goravel"
)

type MinioTestSuite struct {
	suite.Suite
	mockConfig *configmock.Config
	docker     contractsdocker.ImageDriver
	minio      *Minio
}

func TestMinioTestSuite(t *testing.T) {
	suite.Run(t, new(MinioTestSuite))
}

func (s *MinioTestSuite) SetupSuite() {
	s.Nil(os.WriteFile("test.txt", []byte("Goravel"), 0644))

	s.mockConfig = configmock.NewConfig(s.T())
	docker := testingdocker.NewImageDriver(contractsdocker.Image{
		Repository: "minio/minio",
		Tag:        "latest",
		Cmd:        []string{"server", "/data"},
		Env: []string{
			"MINIO_ACCESS_KEY=" + testKey,
			"MINIO_SECRET_KEY=" + testSecret,
		},
		ExposedPorts: []string{
			"9000",
		},
	})
	err := docker.Build()
	if err != nil {
		panic(err)
	}

	config := docker.Config()
	endpoint := fmt.Sprintf("127.0.0.1:%s", supportdocker.ExposedPort(config.ExposedPorts, "9000"))

	if err := docker.Ready(func() error {
		client, err := minio.New(endpoint, &minio.Options{
			Creds: credentials.NewStaticV4(testKey, testSecret, ""),
		})
		if err != nil {
			return err
		}
		if err := client.MakeBucket(context.Background(), testBucket, minio.MakeBucketOptions{}); err != nil {
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
                    "arn:aws:s3:::` + testBucket + `/*"
                ]
            },
            {
                "Action": [
                    "s3:ListBucket"
                ],
                "Effect": "Allow",
                "Principal": "*",
                "Resource": [
                    "arn:aws:s3:::` + testBucket + `"
                ]
            }
        ]
    }`

		if err := client.SetBucketPolicy(context.Background(), testBucket, policy); err != nil {
			return err
		}

		return nil
	}); err != nil {
		panic(err)
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(testKey, testSecret, ""),
		Secure: false,
		Region: "",
	})
	if err != nil {
		panic(err)
	}

	s.docker = docker
	s.minio = &Minio{
		config:   s.mockConfig,
		ctx:      context.Background(),
		instance: client,
		bucket:   testBucket,
		disk:     "minio",
		url:      fmt.Sprintf("http://%s/%s", endpoint, testBucket),
		timezone: "UTC",
	}
}

func (s *MinioTestSuite) TearDownSuite() {
	s.NoError(os.Remove("test.txt"))
	s.NoError(s.docker.Shutdown())
}

func (s *MinioTestSuite) SetupTest() {
}

func (s *MinioTestSuite) TestAllDirectories() {
	s.Nil(s.minio.Put("AllDirectories/1.txt", "Goravel"))
	s.Nil(s.minio.Put("AllDirectories/2.txt", "Goravel"))
	s.Nil(s.minio.Put("AllDirectories/3/3.txt", "Goravel"))
	s.Nil(s.minio.Put("AllDirectories/3/5/6/6.txt", "Goravel"))
	s.Nil(s.minio.MakeDirectory("AllDirectories/3/4"))
	s.True(s.minio.Exists("AllDirectories/1.txt"))
	s.True(s.minio.Exists("AllDirectories/2.txt"))
	s.True(s.minio.Exists("AllDirectories/3/3.txt"))
	s.True(s.minio.Exists("AllDirectories/3/4/"))
	s.True(s.minio.Exists("AllDirectories/"))
	s.True(s.minio.Exists("AllDirectories/3/"))
	s.True(s.minio.Exists("AllDirectories/3/5/"))
	s.True(s.minio.Exists("AllDirectories/3/5/6/"))
	s.True(s.minio.Exists("AllDirectories/3/5/6/6.txt"))
	files, err := s.minio.AllDirectories("AllDirectories")
	s.Nil(err)
	s.Equal([]string{"3/", "3/4/", "3/5/", "3/5/6/"}, files)
	files, err = s.minio.AllDirectories("./AllDirectories")
	s.Nil(err)
	s.Equal([]string{"3/", "3/4/", "3/5/", "3/5/6/"}, files)
	files, err = s.minio.AllDirectories("/AllDirectories")
	s.Nil(err)
	s.Equal([]string{"3/", "3/4/", "3/5/", "3/5/6/"}, files)
	files, err = s.minio.AllDirectories("./AllDirectories/")
	s.Nil(err)
	s.Equal([]string{"3/", "3/4/", "3/5/", "3/5/6/"}, files)
	s.Nil(s.minio.DeleteDirectory("AllDirectories"))
}

func (s *MinioTestSuite) TestAllFiles() {
	s.Nil(s.minio.Put("AllFiles/1.txt", "Goravel"))
	s.Nil(s.minio.Put("AllFiles/2.txt", "Goravel"))
	s.Nil(s.minio.Put("AllFiles/3/3.txt", "Goravel"))
	s.Nil(s.minio.Put("AllFiles/3/4/4.txt", "Goravel"))
	s.True(s.minio.Exists("AllFiles/1.txt"))
	s.True(s.minio.Exists("AllFiles/2.txt"))
	s.True(s.minio.Exists("AllFiles/3/3.txt"))
	s.True(s.minio.Exists("AllFiles/3/4/4.txt"))
	files, err := s.minio.AllFiles("AllFiles")
	s.Nil(err)
	s.Equal([]string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files)
	files, err = s.minio.AllFiles("./AllFiles")
	s.Nil(err)
	s.Equal([]string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files)
	files, err = s.minio.AllFiles("/AllFiles")
	s.Nil(err)
	s.Equal([]string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files)
	files, err = s.minio.AllFiles("./AllFiles/")
	s.Nil(err)
	s.Equal([]string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files)
	s.Nil(s.minio.DeleteDirectory("AllFiles"))
}

func (s *MinioTestSuite) TestCopy() {
	s.Nil(s.minio.Put("Copy/1.txt", "Goravel"))
	s.True(s.minio.Exists("Copy/1.txt"))
	s.Nil(s.minio.Copy("Copy/1.txt", "Copy1/1.txt"))
	s.True(s.minio.Exists("Copy/1.txt"))
	s.True(s.minio.Exists("Copy1/1.txt"))
	s.Nil(s.minio.DeleteDirectory("Copy"))
	s.Nil(s.minio.DeleteDirectory("Copy1"))
}

func (s *MinioTestSuite) TestDelete() {
	s.Nil(s.minio.Put("Delete/1.txt", "Goravel"))
	s.True(s.minio.Exists("Delete/1.txt"))
	s.Nil(s.minio.Delete("Delete/1.txt"))
	s.True(s.minio.Missing("Delete/1.txt"))
	s.Nil(s.minio.DeleteDirectory("Delete"))
}

func (s *MinioTestSuite) TestDeleteDirectory() {
	s.Nil(s.minio.Put("DeleteDirectory/1.txt", "Goravel"))
	s.True(s.minio.Exists("DeleteDirectory/1.txt"))
	s.Nil(s.minio.DeleteDirectory("DeleteDirectory"))
	s.True(s.minio.Missing("DeleteDirectory/1.txt"))
	s.Nil(s.minio.DeleteDirectory("DeleteDirectory"))
}

func (s *MinioTestSuite) TestDirectories() {
	s.Nil(s.minio.Put("Directories/1.txt", "Goravel"))
	s.Nil(s.minio.Put("Directories/2.txt", "Goravel"))
	s.Nil(s.minio.Put("Directories/3/3.txt", "Goravel"))
	s.Nil(s.minio.Put("Directories/3/5/5.txt", "Goravel"))
	s.Nil(s.minio.MakeDirectory("Directories/3/4"))
	s.True(s.minio.Exists("Directories/1.txt"))
	s.True(s.minio.Exists("Directories/2.txt"))
	s.True(s.minio.Exists("Directories/3/3.txt"))
	s.True(s.minio.Exists("Directories/3/4/"))
	s.True(s.minio.Exists("Directories/3/5/5.txt"))
	files, err := s.minio.Directories("Directories")
	s.Nil(err)
	s.Equal([]string{"3/"}, files)
	files, err = s.minio.Directories("./Directories")
	s.Nil(err)
	s.Equal([]string{"3/"}, files)
	files, err = s.minio.Directories("/Directories")
	s.Nil(err)
	s.Equal([]string{"3/"}, files)
	files, err = s.minio.Directories("./Directories/")
	s.Nil(err)
	s.Equal([]string{"3/"}, files)
	s.Nil(s.minio.DeleteDirectory("Directories"))
}

func (s *MinioTestSuite) TestFiles() {
	s.Nil(s.minio.Put("Files/1.txt", "Goravel"))
	s.Nil(s.minio.Put("Files/2.txt", "Goravel"))
	s.Nil(s.minio.Put("Files/3/3.txt", "Goravel"))
	s.Nil(s.minio.Put("Files/3/4/4.txt", "Goravel"))
	s.True(s.minio.Exists("Files/1.txt"))
	s.True(s.minio.Exists("Files/2.txt"))
	s.True(s.minio.Exists("Files/3/3.txt"))
	s.True(s.minio.Exists("Files/3/4/4.txt"))
	files, err := s.minio.Files("Files")
	s.Nil(err)
	s.Equal([]string{"1.txt", "2.txt"}, files)
	files, err = s.minio.Files("./Files")
	s.Nil(err)
	s.Equal([]string{"1.txt", "2.txt"}, files)
	files, err = s.minio.Files("/Files")
	s.Nil(err)
	s.Equal([]string{"1.txt", "2.txt"}, files)
	files, err = s.minio.Files("./Files/")
	s.Nil(err)
	s.Equal([]string{"1.txt", "2.txt"}, files)
	s.Nil(s.minio.DeleteDirectory("Files"))
}

func (s *MinioTestSuite) TestGet() {
	s.Nil(s.minio.Put("Get/1.txt", "Goravel"))
	s.True(s.minio.Exists("Get/1.txt"))
	data, err := s.minio.Get("Get/1.txt")
	s.Nil(err)
	s.Equal("Goravel", data)
	length, err := s.minio.Size("Get/1.txt")
	s.Nil(err)
	s.Equal(int64(7), length)
	s.Nil(s.minio.DeleteDirectory("Get"))
}

func (s *MinioTestSuite) TestGetBytes() {
	s.Nil(s.minio.Put("GetBytes/1.txt", "Goravel"))
	s.True(s.minio.Exists("GetBytes/1.txt"))
	data, err := s.minio.GetBytes("GetBytes/1.txt")
	s.Nil(err)
	s.Equal([]byte("Goravel"), data)
	length, err := s.minio.Size("GetBytes/1.txt")
	s.Nil(err)
	s.Equal(int64(7), length)
	s.Nil(s.minio.DeleteDirectory("GetBytes"))
}

func (s *MinioTestSuite) TestLastModified() {
	s.Nil(s.minio.Put("LastModified/1.txt", "Goravel"))
	s.True(s.minio.Exists("LastModified/1.txt"))
	date, err := s.minio.LastModified("LastModified/1.txt")
	s.Nil(err)

	l, err := time.LoadLocation("UTC")
	s.Nil(err)
	s.Equal(time.Now().In(l).Format("2006-01-02 15"), date.Format("2006-01-02 15"))
	s.Nil(s.minio.DeleteDirectory("LastModified"))
}

func (s *MinioTestSuite) TestMakeDirectory() {
	s.Nil(s.minio.MakeDirectory("MakeDirectory1/"))
	s.Nil(s.minio.MakeDirectory("MakeDirectory2"))
	s.Nil(s.minio.MakeDirectory("MakeDirectory3/MakeDirectory4"))
	s.Nil(s.minio.DeleteDirectory("MakeDirectory1"))
	s.Nil(s.minio.DeleteDirectory("MakeDirectory2"))
	s.Nil(s.minio.DeleteDirectory("MakeDirectory3"))
	s.Nil(s.minio.DeleteDirectory("MakeDirectory4"))
}

func (s *MinioTestSuite) TestMimeType() {
	s.Nil(s.minio.Put("MimeType/1.txt", "Goravel"))
	s.True(s.minio.Exists("MimeType/1.txt"))
	mimeType, err := s.minio.MimeType("MimeType/1.txt")
	s.Nil(err)
	mediaType, _, err := mime.ParseMediaType(mimeType)
	s.Nil(err)
	s.Equal("text/plain", mediaType)

	fileInfo := &File{path: "logo.png"}
	path, err := s.minio.PutFile("MimeType", fileInfo)
	s.Nil(err)
	s.True(s.minio.Exists(path))
	mimeType, err = s.minio.MimeType(path)
	s.Nil(err)
	s.Equal("image/png", mimeType)
}

func (s *MinioTestSuite) TestMove() {
	s.Nil(s.minio.Put("Move/1.txt", "Goravel"))
	s.True(s.minio.Exists("Move/1.txt"))
	s.Nil(s.minio.Move("Move/1.txt", "Move1/1.txt"))
	s.True(s.minio.Missing("Move/1.txt"))
	s.True(s.minio.Exists("Move1/1.txt"))
	s.Nil(s.minio.DeleteDirectory("Move"))
	s.Nil(s.minio.DeleteDirectory("Move1"))
}

func (s *MinioTestSuite) TestPut() {
	s.Nil(s.minio.Put("Put/a/b/1.txt", "Goravel"))
	s.True(s.minio.Exists("Put/"))
	s.True(s.minio.Exists("Put/a/"))
	s.True(s.minio.Exists("Put/a/b/"))
	s.True(s.minio.Exists("Put/a/b/1.txt"))
	s.True(s.minio.Missing("Put/2.txt"))
	s.Nil(s.minio.DeleteDirectory("Put"))
}

func (s *MinioTestSuite) TestPutFile_Image() {
	fileInfo := &File{path: "logo.png"}
	path, err := s.minio.PutFile("PutFile1", fileInfo)
	s.Nil(err)
	s.True(s.minio.Exists(path))
	s.Nil(s.minio.DeleteDirectory("PutFile1"))
}

func (s *MinioTestSuite) TestPutFile_Text() {
	fileInfo := &File{path: "test.txt"}
	path, err := s.minio.PutFile("PutFile", fileInfo)
	s.Nil(err)
	s.True(s.minio.Exists("PutFile/"))
	s.True(s.minio.Exists(path))
	data, err := s.minio.Get(path)
	s.Nil(err)
	s.Equal("Goravel", data)
	s.Nil(s.minio.DeleteDirectory("PutFile"))
}

func (s *MinioTestSuite) TestPutFileAs_Text() {
	fileInfo := &File{path: "test.txt"}
	path, err := s.minio.PutFileAs("PutFileAs", fileInfo, "text")
	s.Nil(err)
	s.Equal("PutFileAs/text.txt", path)
	s.True(s.minio.Exists(path))
	data, err := s.minio.Get(path)
	s.Nil(err)
	s.Equal("Goravel", data)

	path, err = s.minio.PutFileAs("PutFileAs", fileInfo, "text1.txt")
	s.Nil(err)
	s.Equal("PutFileAs/text1.txt", path)
	s.True(s.minio.Exists(path))
	data, err = s.minio.Get(path)
	s.Nil(err)
	s.Equal("Goravel", data)

	s.Nil(s.minio.DeleteDirectory("PutFileAs"))
}

func (s *MinioTestSuite) TestPutFileAs_Image() {
	fileInfo := &File{path: "logo.png"}
	path, err := s.minio.PutFileAs("PutFileAs1", fileInfo, "image")
	s.Nil(err)
	s.Equal("PutFileAs1/image.png", path)
	s.True(s.minio.Exists(path))

	path, err = s.minio.PutFileAs("PutFileAs1", fileInfo, "image1.png")
	s.Nil(err)
	s.Equal("PutFileAs1/image1.png", path)
	s.True(s.minio.Exists(path))

	s.Nil(s.minio.DeleteDirectory("PutFileAs1"))
}

func (s *MinioTestSuite) TestSize() {
	s.Nil(s.minio.Put("Size/1.txt", "Goravel"))
	s.True(s.minio.Exists("Size/1.txt"))
	length, err := s.minio.Size("Size/1.txt")
	s.Nil(err)
	s.Equal(int64(7), length)
	s.Nil(s.minio.DeleteDirectory("Size"))
}

func (s *MinioTestSuite) TestTemporaryUrl() {
	s.Nil(s.minio.Put("TemporaryUrl/1.txt", "Goravel"))
	s.True(s.minio.Exists("TemporaryUrl/1.txt"))
	url, err := s.minio.TemporaryUrl("TemporaryUrl/1.txt", time.Now().Add(5*time.Second))
	s.Nil(err)
	s.NotEmpty(url)
	resp, err := http.Get(url)
	s.Nil(err)
	content, err := io.ReadAll(resp.Body)
	s.Nil(resp.Body.Close())
	s.Nil(err)
	s.Equal("Goravel", string(content))
	s.Nil(s.minio.DeleteDirectory("TemporaryUrl"))
}

func (s *MinioTestSuite) TestUrl() {
	s.Nil(s.minio.Put("Url/1.txt", "Goravel"))
	s.True(s.minio.Exists("Url/1.txt"))
	url := s.minio.url + "/Url/1.txt"
	s.Equal(url, s.minio.Url("Url/1.txt"))
	resp, err := http.Get(url)
	s.Nil(err)
	content, err := io.ReadAll(resp.Body)
	s.Nil(resp.Body.Close())
	s.Nil(err)
	s.Equal("Goravel", string(content))
	s.Nil(s.minio.DeleteDirectory("Url"))
}

type File struct {
	path string
}

func (f *File) Disk(disk string) filesystemcontract.File {
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
