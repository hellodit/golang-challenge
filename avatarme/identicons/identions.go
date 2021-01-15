package identicons

import (
	"crypto/md5"
	"github.com/llgcode/draw2d/draw2dimg"
	"image"
	"image/color"
)

type GridPoint struct {
	value byte
	index int
}

// Identicon Define struct
type Identicon struct {
	Name       string
	Hash       [16]byte
	Color      [3]byte
	grid       []byte
	gridPoints []GridPoint    // Filtered points in the grid
	pixelMap   []DrawingPoint // pixelMap for drawing
}

type Point struct {
	x, y int
}

type DrawingPoint struct {
	topLeft     Point
	bottomRight Point
}

//  HashInput string
func HashInput(input []byte) Identicon {
	// generate checksum from input
	checkSum := md5.Sum(input)
	// return the identicon

	return Identicon{
		Name: string(input),
		Hash: checkSum,
	}
}

func PickColor(identicon Identicon) Identicon {
	// first we make a byte array with length 3
	rgb := [3]byte{}
	// next we copy the first 3 values from the hash to the rgb array
	copy(rgb[:], identicon.Hash[:3])
	// we than assign it to the color value
	identicon.Color = rgb

	return identicon
}

func buildGrind(identicon Identicon) Identicon {
	// Empty frid
	grid := []byte{}
	// Loop over the hash from the identicon
	// Increment with 3 (Chunk the array in 3 parts)
	// this ensures we wont get array out of bounds error and will retrieve exactly 5 chunks of 3
	for i := 0; i < len(identicon.Hash) && i+3 <= len(identicon.Hash)-1; i += 1 {
		// Create a placeholder for the chunk
		chunk := make([]byte, 5)
		copy(chunk, identicon.Hash[i:i+3])
		chunk[3] = chunk[1]           // mirror the second value in the chunk
		chunk[4] = chunk[0]           // mirror the first value in the chunk
		grid = append(grid, chunk...) // append the chunk to the grid

	}
	identicon.grid = grid
	return identicon
}

func filterOddSquares(identicon Identicon) Identicon {
	var grid []GridPoint
	for i, code := range identicon.grid {
		if code%2 == 0 {
			point := GridPoint{
				value: code,
				index: i,
			}
			// append the item to the new grid
			grid = append(grid, point)
		}
	}
	// set the property
	identicon.gridPoints = grid
	return identicon // return the modified identicon
}

func buildPixelMap(identicon Identicon) Identicon {
	var drawingPoints []DrawingPoint // define placeholder for drawingpoints
	// Closure, this function returns a Drawingpoint

	pixelFunc := func(p GridPoint) DrawingPoint {
		horizontal := (p.index % 5) * 50
		vertical := (p.index / 5) * 50
		// this is the topleft point with x and the y
		topLeft := Point{horizontal, vertical}
		// the bottom right point is just the topleft point +50 because 1 block in the grid is 50x50
		bottomRight := Point{horizontal + 50, vertical + 50}

		return DrawingPoint{ // We then return the drawingpoint
			topLeft,
			bottomRight,
		}
	}

	for _, gridPoint := range identicon.gridPoints {
		// for every gridPoint we calculate the drawingpoints and we add them to the array
		drawingPoints = append(drawingPoints, pixelFunc(gridPoint))
	}
	identicon.pixelMap = drawingPoints // set the drawingpoint value on the identicon
	return identicon                   // return the modified identicon
}

func generateRec(img *image.RGBA, col color.Color, x1, y1, x2, y2 float64) {
	gc := draw2dimg.NewGraphicContext(img)
	gc.SetFillColor(col) // set the color
	gc.MoveTo(x1, y1)    // move to the topleft in the image
	// Draw the lines for the dimensions
	gc.LineTo(x1, y1)
	gc.LineTo(x1, y2)
	gc.MoveTo(x2, y1) // move to the right in the image
	// Draw the lines for the dimensions
	gc.LineTo(x2, y1)
	gc.LineTo(x2, y2)
	// Set the linewidth to zero
	gc.SetLineWidth(0)
	// Fill the stroke so the rectangle will be filled
	gc.FillStroke()
}

func drawRectangle(identicon Identicon) error {
	// We create our default image containing a 250x250 rectangle
	var img = image.NewRGBA(image.Rect(0, 0, 250, 250))
	// We retrieve the color from the color property on the identicon
	col := color.RGBA{R: identicon.Color[0], G: identicon.Color[1], B: identicon.Color[2], A: 255}

	// Loop over the pixelmap and call the rect function with the img, color and the dimensions
	for _, pixel := range identicon.pixelMap {
		generateRec(
			img,
			col,
			float64(pixel.topLeft.x),
			float64(pixel.topLeft.y),
			float64(pixel.bottomRight.x),
			float64(pixel.bottomRight.y),
		)
	}
	// Finally save the image to disk
	return draw2dimg.SaveToPngFile(identicon.Name+".png", img)
}

type Apply func(Identicon) Identicon

func Pipe(identicon Identicon, funcs ...Apply) Identicon {
	for _, applyer := range funcs {
		identicon = applyer(identicon)
	}
	return identicon
}
