// This program generates .nptl mesh parametrically
// The output file is on stdout.
// .nptl can be converted to .ply using PoissonRecon
// http://research.microsoft.com/en-us/um/people/hoppe/proj/poissonrecon/
// .ply can be converted to stl, using meshlab
package main

import (
	"fmt"
	"image"
	"image/color"
	//"image/png"
	"log"
	"math"
	"os"

	"github.com/krasin/voxel/nptl"
	"github.com/krasin/voxel/raster"
	"github.com/krasin/voxel/volume"
)

const (
	VoxelSide = 512

	eps = 1E-4
)

var (
	_ = fmt.Printf
)

type Paramesh interface {
	V3(t float64) [3]float64
	V2(t float64) [2]float64
	S1(t float64) float64
}

type paramCube struct {
	Side float64
}

func (cube *paramCube) S1(t float64) float64    { return cube.Side }
func (cube *paramCube) V2(t float64) [2]float64 { return [2]float64{cube.Side, 0} }
func (cube *paramCube) V3(t float64) [3]float64 { return [3]float64{cube.Side, 0, 0} }

type triangle struct {
	Side float64
}

func (cube *triangle) S1(t float64) float64    { return cube.Side * t }
func (cube *triangle) V2(t float64) [2]float64 { return [2]float64{cube.Side, 0} }
func (cube *triangle) V3(t float64) [3]float64 { return [3]float64{cube.Side, 0, 0} }

type circle struct {
	Side float64
}

func (cube *circle) S1(t float64) float64 { return 0.4 * cube.Side } // * math.Sqrt(1-4*(t-0.5)*(t-0.5)) }
func (cube *circle) V2(t float64) [2]float64 {
	return [2]float64{
		2 * cube.Side * math.Sin(2*t*math.Pi),
		2 * cube.Side * math.Cos(2*t*math.Pi),
	}
}
func (cube *circle) V3(t float64) [3]float64 { return [3]float64{cube.Side, 0, 0} }

func norm(v [3]float64) float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
}

func norm2(v [2]float64) float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1])
}

func normalize(v [3]float64) [3]float64 {
	l := norm(v)
	if l < eps {
		return v
	}
	return [3]float64{
		v[0] / l,
		v[1] / l,
		v[2] / l,
	}
}

func normalize2(v [2]float64) [2]float64 {
	l := norm2(v)
	if l < eps {
		return v
	}
	return [2]float64{
		v[0] / l,
		v[1] / l,
	}
}

func findXY(v [3]float64) (x, y [3]float64) {
	v = normalize(v)
	tr := [3]int{0, 1, 2}
	if math.Abs(v[1]) > math.Abs(v[0]) {
		tr[0], tr[1] = tr[1], tr[0]
	}
	if math.Abs(v[2]) > math.Abs(v[tr[0]]) {
		tr[0], tr[2] = tr[2], tr[0]
	}
	x = normalize([3]float64{-v[tr[1]] / v[tr[0]], 1, 0})
	y = normalize([3]float64{-v[tr[2]] / v[tr[0]], 0, 1})
	return
}

func findX(v [2]float64) [2]float64 {
	return normalize2([2]float64{v[1], -v[0]})
	/*	tr := [2]int{0, 1}
		if math.Abs(v[1]) > math.Abs(v[0]) {
			tr[0], tr[1] = tr[1], tr[0]
		}
		xx := normalize([3]float64{-v[tr[1]] / v[tr[0]], 1, 0})
		x = [2]float64{xx[0], xx[1]}
		return*/
}

/*func rasterize(mesh Paramesh, vol *volume.SparseVolume, grid *raster.Grid, step float64) {
	var cur [3]float64
	for t3 := float64(0); t3 <= 1; t3 += step {
		v3 := mesh.V3(t3)
		next3 := [3]float64{
			cur[0] + v3[0]*step,
			cur[1] + v3[1]*step,
			cur[2] + v3[2]*step,
		}
		x2, y2 := findXY(v3)
		var cur2 [2]float64
		for t2 := float64(0); t2 <= 1; t2 += step {
			v2 := mesh.V2(t2)
			next2 := [3]float64{
				cur2[0] + v2[0]*step,
				cur2[1] + v2[1]*step,
			}
			x1 := findX(v2)
			for t1 := float64(0); t1 <= mesh.S1(t2); t1 += step {
				p := [3]float64{
					v3[0] + (v2[0]+x1[0]*t1)*x2[0] + (v2[1]+x1[1]*t1)*y2[0],
					v3[1] + (v2[0]+x1[0]*t1)*x2[1] + (v2[1]+x1[1]*t1)*y2[1],
					v3[2] + (v2[0]+x1[0]*t1)*x2[2] + (v2[1]+x1[1]*t1)*y2[2],
				}
				vol.Set(int(float64(grid.N[0])*(p[0]-grid.P0[0])/(grid.P1[0]-grid.P0[0])),
					int(float64(grid.N[1])*(p[1]-grid.P0[1])/(grid.P1[1]-grid.P0[1])),
					int(float64(grid.N[2])*(p[2]-grid.P0[2])/(grid.P1[2]-grid.P0[2])),
					1)
			}
		}
	}
}*/

func rotateXY(p [3]float64, alpha float64) [3]float64 {
	return [3]float64{
		p[0]*math.Cos(alpha) + p[1]*math.Sin(alpha),
		-p[0]*math.Sin(alpha) + p[1]*math.Cos(alpha),
		p[2],
	}
}

func rotateXZ(p [3]float64, alpha float64) [3]float64 {
	return [3]float64{
		p[0]*math.Cos(alpha) + p[2]*math.Sin(alpha),
		p[1],
		-p[0]*math.Sin(alpha) + p[2]*math.Cos(alpha),
	}
}

func rotateYZ(p [3]float64, alpha float64) [3]float64 {
	return [3]float64{
		p[0],
		p[1]*math.Cos(alpha) + p[2]*math.Sin(alpha),
		-p[1]*math.Sin(alpha) + p[2]*math.Cos(alpha),
	}
}

func rotate3(p, angles [3]float64) [3]float64 {
	p = rotateXY(p, angles[0])
	p = rotateXZ(p, angles[1])
	p = rotateYZ(p, angles[2])
	return p
}

func draw3d(mesh Paramesh, vol *volume.SparseVolume, grid *raster.Grid, step float64) {
	var cur [3]float64
	front := [3]float64{1, 0, 0}
	side := [3]float64{0, 1, 0}
	up := [3]float64{0, 0, 1}

	var t3 float64
	setPixel := func(x, y float64) {
		//		fmt.Printf("setPixel(x=%f, y=%f) ", x, y)
		p := [3]float64{
			cur[0] + side[0]*x + up[0]*y,
			cur[1] + side[1]*x + up[1]*y,
			cur[2] + side[2]*x + up[2]*y,
		}
		p2 := [3]int{
			int(float64(grid.N[0]) * (p[0] - grid.P0[0]) / (grid.P1[0] - grid.P0[0])),
			int(float64(grid.N[1]) * (p[1] - grid.P0[1]) / (grid.P1[1] - grid.P0[1])),
			int(float64(grid.N[2]) * (p[2] - grid.P0[2]) / (grid.P1[2] - grid.P0[2])),
		}
		//		fmt.Printf("vol.Set(%d, %d, %d, 1)\n", p2[0], p2[1], p2[2])
		vol.Set(p2[0], p2[1], p2[2], 1)
	}
	for t3 = 0; t3 <= 1; t3 += step {
		v3 := mesh.V3(t3)
		draw2d2(mesh, setPixel, step)
		cur = [3]float64{
			cur[0] + front[0]*step,
			cur[1] + front[1]*step,
			cur[2] + front[2]*step,
		}
		front = rotate3(front, v3)
		side = rotate3(side, v3)
		up = rotate3(up, v3)
	}
}

func draw2d2(mesh Paramesh, setPixel func(x, y float64), step float64) {
	var cur [2]float64
	for t2 := float64(0); t2 <= 1; t2 += step {
		v2 := mesh.V2(t2)

		x1 := findX(v2)

		len1 := mesh.S1(t2)
		for t1 := float64(0); t1 <= 1; t1 += step {
			cur1 := t1 * len1
			setPixel(cur[0]+cur1*x1[0],
				cur[1]+cur1*x1[1])
		}
		cur[0] += v2[0] * step
		cur[1] += v2[1] * step
	}
}

func draw2d(x0, y0 int, mesh Paramesh, img *image.RGBA, step float64) {
	var cur [2]float64
	for t2 := float64(0); t2 <= 1; t2 += step {
		v2 := mesh.V2(t2)

		x1 := findX(v2)

		len1 := mesh.S1(t2)
		for t1 := float64(0); t1 <= 1; t1 += step {
			cur1 := t1 * len1

			// p := (cur + cur1 * x1)*img.Bounds
			p := [2]int{
				x0 + int((cur[0] + cur1*x1[0])),
				y0 + int((cur[1] + cur1*x1[1])),
			}
			//			fmt.Printf("%v", p)
			img.Set(p[0], p[1], color.RGBA{0, 0, 255, 255})
		}

		cur[0] += v2[0] * step
		cur[1] += v2[1] * step
	}
}

func main() {
	//	img := image.NewRGBA(image.Rect(0, 0, 1024, 768))
	//	draw2d(100, 100, &paramCube{500}, img, 0.002)
	//	draw2d(100, 100, &triangle{500}, img, 0.002)
	//	draw2d(100, 100, &circle{300}, img, 0.002)
	//	f, err := os.OpenFile("lala.png", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	defer f.Close()
	//	if err = png.Encode(f, img); err != nil {
	//		log.Fatal(err)
	//	}

	vol := volume.NewSparseVolume(VoxelSide)

	grid := raster.Grid{
		P0: [3]float64{-512, -512, -512},
		P1: [3]float64{512, 512, 512},
		N:  [3]int64{VoxelSide, VoxelSide, VoxelSide},
	}

	draw3d(&paramCube{256}, vol, &grid, 0.001)

	grid2 := raster.Grid{
		P0: [3]float64{0, 0, 0},
		P1: [3]float64{1024, 1024, 1024},
		N:  [3]int64{VoxelSide, VoxelSide, VoxelSide},
	}

	if err := nptl.WriteNptl(vol, grid2, os.Stdout); err != nil {
		log.Fatalf("WriteNptl: %v", err)
	}

}
