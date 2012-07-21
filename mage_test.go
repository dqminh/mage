package mage

import (
  . "launchpad.net/gocheck"
  "io/ioutil"
  "os"
  "testing"
  "image"
  _ "image/jpeg"
  _ "image/png"
  _ "image/gif"
)

func Test(t *testing.T) { TestingT(t) }
type S struct{}
var _ = Suite(&S{})

func (s *S) SetUpTest(c *C) {
  InitWandEnv()
}

func (s *S) TearDownTest(c *C) {
  TermWandEnv()
}

func assertSize(c *C, image image.Image, x int, y int) {
  bounds := image.Bounds()
  c.Assert(bounds.Dx(), Equals, x)
  c.Assert(bounds.Dy(), Equals, y)
}

func readImage(c *C, filename string) *Mage {
  var im *Mage;
  original, err := ioutil.ReadFile(filename)
  c.Check(err, IsNil, Commentf("Fail to read image file"))

  im = NewMage()
  success := im.ReadBlob(original)
  c.Check(success, Equals, true, Commentf("Fail to read blob"))
  return im
}

func writeFile(im *Mage, filename string) {
  ioutil.WriteFile(filename, im.ExportBlob(), 0644)
}

func (s *S) TestReadAndExportBlob(c *C) {
  filename := "images/out/test_read_and_export_blob.jpg"
  im := readImage(c, "images/in/test.jpg")
  writeFile(im, filename)
  out, _ := os.Open(filename)
  defer out.Close()
  exported, _, _ := image.Decode(out)
  assertSize(c, exported, 500, 371)
}

func (s *S) TestWidth(c *C) {
  im := readImage(c, "images/in/test.jpg")
  c.Assert(im.Width(), Equals, 500)
}

func (s *S) TestHeight(c *C) {
  im := readImage(c, "images/in/test.jpg")
  c.Assert(im.Height(), Equals, 371)
}

func (s *S) TestResize(c *C) {
  filename := "images/out/test_resize.jpg"
  im := readImage(c, "images/in/test.jpg")
  expectedWidth := 100
  expectedHeight := 100
  im.Resize(expectedWidth, expectedHeight)
  writeFile(im, filename)
  out, _ := os.Open(filename)
  defer out.Close()
  exported, _, _ := image.Decode(out)
  assertSize(c, exported, expectedWidth, expectedHeight)
}
