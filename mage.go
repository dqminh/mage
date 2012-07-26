package mage

// #cgo pkg-config: MagickWand MagickCore
// #include <stdlib.h>
// #include <wand/MagickWand.h>
import "C"

import (
  "unsafe"
  "math"
)

type Mage struct {
  wand *C.MagickWand
}

// Private: Convert C.MagickBooleanType to boolean type
//
// Params:
//  - b: either C.MagickTrue or C.MagickFalse
//
// Examples
//  mBoolean(C.MagickTrue) == true
//  mBoolean(C.MagickFalse) == false
func mBoolean(b C.MagickBooleanType) bool {
  return b == C.MagickTrue
}

// Private: Round up a float64 number to int
//
// Examples
//  round(float64(0.5)) == 1
//  round(float64(0.6)) == 1
//  round(float64(0.3)) == 0
func round(x float64) int {
  return int(math.Floor(x + 0.5))
}

// Private: Create a blank magick wand with size width and height
//
// Params:
// - format: format of the new image
// - width: width of the new image
// - height: height of the new image
//
// Examples
//  blankWand("jpg", 100, 100)
//
// Return *C.MagickWand
func blankWand(format string, width, height int) *C.MagickWand {
  wand := C.NewMagickWand()
  cformat := C.CString(format)
  noneBackground := C.CString("none")
  defer C.free(unsafe.Pointer(cformat))
  defer C.free(unsafe.Pointer(noneBackground))

  C.MagickSetFormat(wand, C.CString(format))
  pixel := C.NewPixelWand()
  defer C.DestroyPixelWand(pixel)

  C.PixelSetColor(pixel, noneBackground);
  C.MagickSetSize(wand, C.ulong(width), C.ulong(height));
  C.MagickNewImage(wand, C.ulong(width), C.ulong(height), pixel)
  return wand
}

// Private: scale wand's image respected the initial ratio. At least one of the
// new dimension will be equal to one of the passed in width/height. This
// depends on the maximum scale between the old dimension and desired dimension
//
// Params:
// - width: probably new width of the scaled image
// - height: probably new height of the scaled image
//
// Examples: given an image with initial dimention: 1000x1000
//  m.scale(300, 500) == 500, 500
//  m.scale(300, 200) == 300, 300
//
// Returns a pair of scaled width and height of the new image
func (m *Mage) scale(width , height int) (scaledWidth, scaledHeight int) {
  scale := float64(1.0)
  imageWidth := m.Width()
  imageHeight := m.Height()
  if width != imageWidth || height != imageHeight {
    scale = math.Max(
      float64(width)/float64(imageWidth),
      float64(height)/float64(imageHeight))
  }
  scaledWidth = round(scale * (float64(imageWidth) + 0.5))
  scaledHeight = round(scale * (float64(imageHeight) + 0.5))
  return scaledWidth, scaledHeight
}

// Private: center the current wand on top of the new wand, after that, we only
// keep the new wand.
//
// Params:
// - newWand: new wand, probably going to be result of blankWand
// - x: x position on the current wand that is the center of the new wand
// - y: y position on the current wand that is the center of the new wand
//
// Example:
//  newWand := blankWand("jpg", width, height)
//  done := m.compositeCenter(newWand, 10, 10)
//
// Return boolean result of the composition
func (m *Mage) compositeCenter(newWand *C.MagickWand, x, y int) bool{
  success := C.MagickCompositeImage(
    newWand,
    m.wand,
    C.OverCompositeOp,
    C.long(x),
    C.long(y))
  C.DestroyMagickWand(m.wand)
  m.wand = newWand
  return mBoolean(success)
}

// Private: resize the current image to new width and height
//
// Params:
// - width: width of the new image
// - height: height of the new image
//
// Examples:
//  m.resize(100, 100)
func (m *Mage) resize(width, height int) bool {
  return mBoolean(C.MagickResizeImage(
    m.wand,
    C.ulong(width),
    C.ulong(height),
    C.LanczosFilter,
    C.double(1.0)))
}

// Private: strip all comments and profiles from an image
//
// Examples:
//  m.strip()
func (m *Mage) strip() bool {
  return mBoolean(C.MagickStripImage(m.wand))
}

// Public: read a blob data into the current wand
//
// Examples:
//  im = NewMage()
//  original, err := ioutil.ReadFile("test.jpg")
//  success := im.ReadBlob(original)
func (m *Mage) ReadBlob(blob []byte) bool {
  return mBoolean(C.MagickReadImageBlob(
    m.wand,
    unsafe.Pointer(&blob[0]),
    C.ulong(len(blob))))
}

// Public: export the current image into a blob. Also destroy the current wand
//
// Examples:
//  im = NewMage()
//  original, err := ioutil.ReadFile("test.jpg")
//  success := im.ReadBlob(original)
//  imageBytes := im.ExportBlob()
func (m *Mage) ExportBlob() []byte {
  defer m.Destroy()
  newSize := C.ulong(0)
  C.MagickResetIterator(m.wand)
  image := C.MagickGetImageBlob(m.wand, &newSize)
  imagePointer := unsafe.Pointer(image)
  defer C.MagickRelinquishMemory(imagePointer)
  return C.GoBytes(imagePointer, C.int(newSize))
}

// Public: resize an image. The algorithm to resize the image is:
//  - strip all comments and profiles data
//  - resize the image respect to the original ratio
//  - center the image with the new size, remove anything that isnt in the new
//  dimension
//
// Params:
//  - width: new width
//  - height: new height
//
// Examples:
//  im = NewMage()
//  original, err := ioutil.ReadFile("test.jpg")
//  success := im.ReadBlob(original)
//  success = im.Resize(100, 100)
func (m *Mage) Resize(width, height int) bool {
  var done bool;
  scaledWidth, scaledHeight := m.scale(width, height)
  done = m.strip()
  done = m.resize(scaledWidth, scaledHeight)
  newWand := blankWand("jpg", width, height)
  done = m.compositeCenter(newWand, int((width - scaledWidth) / 2), int((height - scaledHeight) / 2))
  return done
}

// Public: get current width of the image
//
// Examples:
//  im = NewMage()
//  original, err := ioutil.ReadFile("test.jpg")
//  im.Width()
func (m *Mage) Width() int {
  return int(C.MagickGetImageWidth(m.wand))
}

// Public: get current height of the image
//
// Examples:
//  im = NewMage()
//  original, err := ioutil.ReadFile("test.jpg")
//  im.Height()
func (m *Mage) Height() int {
  return int(C.MagickGetImageHeight(m.wand))
}

func (m *Mage) Destroy() {
  defer C.DestroyMagickWand(m.wand)
}

// Public: initialize the magick wand environment
// This should only be called at the start of the process, before any magick
// operation
func InitWandEnv() {
  C.MagickWandGenesis()
}

// Public: destroy the magick wand environment
// This should only be called at the end of the process, after all magick
// operation
func TermWandEnv() {
  C.MagickWandTerminus()
}

// Public: create a new mage, associate with a new magick wand
//
// Examples:
//  InitWandEnv()
//  mage := NewMage()
//  ...
//  TermWandEnv()
func NewMage() *Mage {
  mage := &Mage{}
  mage.wand = C.NewMagickWand()
  return mage
}
