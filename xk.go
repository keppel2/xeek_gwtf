package main

import "github.com/kshedden/gonpy"
import "fmt"
import "math/rand"
import "math"
import "bytes"
import "encoding/binary"
import "io/ioutil"
import "os"

const XSIZE, YSIZE, ZSIZE = 300, 300, 1250
const SIZE = XSIZE * YSIZE * ZSIZE
const TURNS = 2000
const AGENTSPT = 100
const SPAWN_NEW = true

var RANDOMGEN = 20
var MURD = 0

var CIRC = 10
var RANDOMMOVE = 0

const SEED = 7

func checker(x, y, z, dx, dy, dz int, acc []bool) bool {
	return false

}

func csvStr(x, y, z int) string {
	return fmt.Sprintf("%v,%v,%v\n", x, y, z)

}

func pyramad(x, y, z, tot int) string {
	rt := csvStr(x, y, z)
	count := 1
	if count >= tot {
		return rt
	}
	cd := 0
	for {
		cd++
		for dx := 0; dx <= cd; dx++ {
			for dy := cd - dx; dx+dy <= cd; dy++ {
				dz := cd - dx - dy
				dxs := []int{-dx, dx}
				dys := []int{-dy, dy}
				dzs := []int{-dz, dz}
				for _, cdx := range dxs {
					for _, cdy := range dys {
						for _, cdz := range dzs {
							rt += csvStr(x+cdx, y+cdy, z+cdz)
							count++
							if count >= tot {
								return rt
							}
						}
					}
				}

			}
		}
	}

}

var xs uint64 = SEED

func xshift(xs uint64) (uint64, uint64) {
	xs ^= xs >> 12
	xs ^= xs << 25
	xs ^= xs >> 27
	return xs * 0x2545F4914F6CDD1D, xs
}

var RCOUNT = 0

func randi(mod int) int {
	var x uint64
	x, xs = xshift(xs)
	return int(x % uint64(mod))

	//	return rand.Intn(a)
}
func randit(mod int, t *took) int {
	var x uint64
	x, t.seed = xshift(t.seed)
	return int(x % uint64(mod))
}

type took struct {
	seed uint64
	ind  int

	ar       []bool
	count    int
	ags      []*agent
	accc     [][3]uint16
	envar    []uint8
	rangen   int
	ranmove  int
	prefs    [2][3][3]int
	probs    [2][3][3]int
	cprobs   [][4]int
	tracking []int16
}

var fa [SIZE]bool
var mda [][][]bool

var probs [2][3][3]int = [2][3][3]int{
	{
		{1, 1, 1},
		{1, 0, 1},
		{1, 1, 1}}, {
		{2, 2, 2},
		{2, 3, 2},
		{2, 2, 2}}}

var Gprefs [2][3][3]int
var GprefsN [2][3][3]int = [2][3][3]int{ //835
	{{1, 1, 1},
		{1, 0, 1},
		{1, 1, 1}}, {
		{2, 2, 2},
		{2, 3, 2},
		{2, 2, 2}}}

var GprefsA [2][3][3]int = [2][3][3]int{
	{{1, 1, 1},
		{1, 0, 1},
		{1, 1, 4}}, {
		{2, 2, 2},
		{2, 3, 2},
		{2, 2, 3}}}

var GprefsX [2][3][3]int = GprefsA

//var GprefsX [2][3][3]int = [2][3][3]int{[3][3]int{[3]int{4, 3, 4}, [3]int{3, 0, 3}, [3]int{4, 3, 4}}, [3][3]int{[3]int{2, 3, 2}, [3]int{3, 3, 3}, [3]int{2, 3, 2}}}
//var GprefsX [2][3][3]int = [2][3][3]int{[3][3]int{[3]int{8, 7, 8}, [3]int{7, 0, 7}, [3]int{8, 7, 8}}, [3][3]int{[3]int{2, 3, 2}, [3]int{3, 3, 3}, [3]int{2, 3, 2}}}

var GprefL [][2][3][3]int = [][2][3][3]int{
	{{{1, 1, 1},
		{1, 0, 1},
		{1, 1, 1}}, {
		{2, 2, 2},
		{2, 3, 2},
		{2, 2, 2}}},
	{{{3, 1, 3},
		{1, 0, 1},
		{3, 1, 3}}, {
		{2, 2, 2},
		{2, 3, 2},
		{2, 2, 2}}},

	{{{1, 3, 1},
		{3, 0, 3},
		{1, 3, 1}}, {
		{2, 2, 2},
		{2, 3, 2},
		{2, 2, 2}}},
} /*
	{{{1, 1, 1},
		{1, 0, 1},
		{3, 1, 1}}, {
		{2, 2, 2},
		{2, 3, 2},
		{2, 2, 2}}},
}

/*
[[[1 1 1] [1 0 1] [1 1 6]] [[2 2 2] [2 3 2] [2 2 2]]]
588.0542433580646 xk 46789591.2177
[[[1 1 1] [1 0 1] [1 1 12]] [[2 2 2] [2 3 2] [2 2 2]]]
638.3906188566243

*/

func cprobs(ar [2][3][3]int) [][4]int {
	rt := make([][4]int, 0)
	add := 0
	for a, x := range probs {
		for b, y := range x {
			for c, z := range y {
				if z == 0 {
					continue
				}
				add += z
				tup := [4]int{add, c - 1, b - 1, -a}
				rt = append(rt, tup)
			}
		}
	}
	return rt
}

func (t *took) startUp() {
	t.count = 0
	t.rangen = RANDOMGEN
	t.ranmove = RANDOMMOVE
	t.prefs = Gprefs
	t.probs = probs
	t.cprobs = cprobs(t.probs)
	t.seed = SEED
	randit(1, t)
	randit(1, t)
	/*
	   	   t.tracking = make ([]int16, TURNS * AGENTSPT * TURNS * 3)
	           for x := 0;x < len(t.tracking);x++ {
	                   t.tracking[x] = -1
	           }
	*/
}

func (t took) disp() string {
	rt := fmt.Sprintln("gen", t.rangen, "move", t.ranmove, "turns", TURNS, "ags", AGENTSPT)
	rt += fmt.Sprintln(t.prefs)
	rt += fmt.Sprintln(t.probs)
	return rt
}

func inCol(x, y, z int) bool {
	return x >= 0 && x < XSIZE && y >= 0 && y < YSIZE && z >= 0 && z < ZSIZE

}

func (t took) legal(x, y, z int) bool {
	return inCol(x, y, z) && !t.ar[to3(x, y, z)]
}

func (t *took) add(x, y, z int) {
	t.count++
	t.ar[to3(x, y, z)] = true
}

func (t *took) movePoint(x, y, z, dx, dy, dz int) {
	if t.ar[to3(x, y, z)] == false {
		panic(fmt.Sprintln(x, y, z, dx, dy, dz))
	}
	t.ar[to3(x, y, z)] = false
	if t.ar[to3(x+dx, y+dy, z+dz)] == true {
		panic(7)
	}
	t.ar[to3(x+dx, y+dy, z+dz)] = true
}

type agent struct {
	aid     int
	x, y, z int
	dead    bool
	md      float64
}

func (a *agent) startUp() {
	a.x, a.y, a.z = -1, -1, -1
	a.aid = -1
	a.dead = false
}
func (a *agent) iter(tk *took, t int) {
	/*
	   tk.tracking[to1t(a.aid, t, 0)] = int16(a.x)
	   tk.tracking[to1t(a.aid, t, 1)] = int16(a.y)
	   tk.tracking[to1t(a.aid, t, 2)] = int16(a.z)
	*/

	if a.dead {
		return
	}
	/*
	   if randit(100, tk) < 1 && a.cube(3, *tk) > 32 {
	           a.dead = true
	           return
	   }
	*/
	if !a.hasRoom(*tk) {
		a.dead = true
		return
	}

	if randit(100, tk) < tk.ranmove {
		a.rMove(tk)
	} else {
		a.move(tk)
	}
	/*
	   if randit(1000, tk) < MURD { //&& tk.envar[to3(a.x,a.y,a.z)] < 88 {
	           a.dead = true
	   }
	*/

}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
func absf(a float64) float64 {
	if a < 0 {
		return -a
	}
	return a
}
func dist(ax, ay, az, bx, by, bz int) float64 {
	int := (ax-bx)*(ax-bx) + (ay-by)*(ay-by) + (az-bz)*(az-bz)
	return float64(int)
}

func (a agent) mdist(f float64) float64 {
	return absf(f - a.md)
}

func (a *agent) adist(tk took) float64 {
	rt := -1.0
	for _, acc := range tk.accc {
		x := dist(a.x, a.y, a.z, int(acc[0]), int(acc[1]), int(acc[2]))
		if rt == -1.0 || x < rt {
			rt = x
		}
	}
	a.md = rt
	return rt
}

func (a *agent) exMove(dx, dy, dz int, tk *took) {
	tk.movePoint(a.x, a.y, a.z, dx, dy, dz)
	a.x += dx
	a.y += dy
	a.z += dz
	if a.z == 0 {
		a.dead = true
	}
}

func (a agent) cube(e int, tk took) int {
	count := 0
	for x := a.x - e; x <= a.x+e; x++ {
		for y := a.y - e; y <= a.y+e; y++ {
			for z := a.z - e; z <= a.z+e; z++ {
				if inCol(x, y, z) && tk.ar[to3(x, y, z)] {
					count++
				}
			}
		}
	}
	return count
}

func (a agent) hasRoom(tk took) bool {
	lCount := 0
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			for dz := 0; dz >= -1; dz-- {
				if dz == 0 && dy == 0 && dx == 0 {
					continue
				}
				if tk.legal(a.x+dx, a.y+dy, a.z+dz) {
					lCount++
					return true
				}
			}
		}
	}
	if lCount < 1 {
		return false
	}
	return true
}

func (ag *agent) move(tk *took) {
	maxn := -1
	var maxes [][3]int
	if !ag.hasRoom(*tk) {
		panic(5)
	}
	for a, aa := range tk.prefs {
		for b, bb := range aa {
			for c, pre := range bb {
				dx := c - 1
				dy := b - 1
				dz := -a
				if dx == 0 && dy == 0 && dz == 0 {
					continue
				}
				tx := ag.x + dx
				ty := ag.y + dy
				tz := ag.z + dz
				if !tk.legal(tx, ty, tz) {
					continue
				}
				p := pre
				FAC := 20
				_ = FAC
				/*

				   if ag.x < 150 && dx == 1 {
				           p += (150 - ag.x) / FAC
				   }
				   if ag.x > 150 && dx == -1 {
				           p += (ag.x - 150) / FAC
				   }
				    if ag.y < 150 && dy == 1 {
				           p += (150 - ag.y) / FAC
				   }
				   if ag.y > 150 && dy == -1 {
				           p += (ag.y - 150) / FAC
				   }
				*/

				p +=
					int(tk.envar[to3(tx, ty, tz)])
				if p >= maxn {
					maxn = p
					maxes = append(maxes, [3]int{dx, dy, dz})
				}
			}
		}
	}
	if len(maxes) == 0 {
		panic(7)
	}
	r := randit(len(maxes), tk)
	mymax := maxes[r]
	ag.exMove(mymax[0], mymax[1], mymax[2], tk)
}

func (a *agent) rMove(tk *took) {
	for {
		r := randit(tk.cprobs[len(tk.cprobs)-1][0], tk)
		var cp [4]int
		for _, v := range tk.cprobs {
			if r < v[0] {
				cp = v //cpref[k]
				dx := cp[1]
				dy := cp[2]
				dz := cp[3]
				if tk.legal(a.x+dx, a.y+dy, a.z+dz) {
					a.exMove(dx, dy, dz, tk)
					return
				}
				break
			}
		}
	}
}

func to1t(x, y, z int) int {
	return x*TURNS*3 + y*3 + z
}
func to1(a int) (int, int, int) {
	c := a
	z := c % ZSIZE
	c /= ZSIZE
	y := c % YSIZE
	c /= YSIZE
	x := c
	return x, y, z
}

func to3(x, y, z int) int {
	return x*YSIZE*ZSIZE + y*ZSIZE + z
}

func loadEA(a int) []uint8 {
	rt, _ := ioutil.ReadFile(fmt.Sprintf("vol%venv.dat", a))
	return rt
}
func loadAc(a int) [][3]uint16 {
	ba, _ := ioutil.ReadFile(fmt.Sprintf("acc%v.dat", a))
	buf2 := bytes.NewReader(ba)
	rt := make([][3]uint16, len(ba)/binary.Size([3]uint16{}))
	binary.Read(buf2, binary.LittleEndian, rt)
	return rt
}

func (tk took) score() (float64, float64) {
	sc := 0.0
	mc := make(chan float64)
	cur := 0
	for th := 0; th < TH; th++ {
		if th == TH-1 {
			go gorange(cur, len(tk.ags), tk, mc)
		} else {
			go gorange(cur, cur+len(tk.ags)/TH, tk, mc)
		}
		cur += len(tk.ags) / TH
	}

	for th := 0; th < TH; th++ {
		sc += <-mc
	}
	sc /= TURNS * AGENTSPT
	vari := sc
	sc = math.Sqrt(sc)

	scd := 0.0
	for _, a := range tk.ags {
		scd += a.mdist(vari)
	}
	scd /= TURNS * AGENTSPT
	scd = math.Sqrt(scd)

	return sc, scd
}
func (tk took) statz() (int, int) {
	dead := 0
	top := 0
	for _, an := range tk.ags {
		if an.dead && an.z != 0 {
			dead++
		}
		if an.z == 0 {
			top++
		}
	}
	return dead, top

}
func (tk took) trackp() {

	w, _ := gonpy.NewFileWriter("track6.npy")
	w.Shape = []int{200000, 2000, 3}
	_ = w.WriteInt16(tk.tracking)

}
func (tk took) wnpy() {
	w, _ := gonpy.NewFileWriter("acc.npy")
	w.Shape = []int{300, 300, 1250}
	mar := make([]uint8, XSIZE*YSIZE*ZSIZE)
	for k, v := range tk.ar {
		if v {
			mar[k] = 1
		}
	}
	_ = w.WriteUint8(mar)

}
func (tk took) prn() {
	fmt.Println("x,y,z")
	for _, an := range tk.ags {
		fmt.Printf("%v,%v,%v\n", an.x, an.y, an.z)
	}
}

func (tk *took) surround(x, y, z, tot int) {

	for cL := 0; ; cL++ {
		for tx := 0; tx < XSIZE; tx++ {
			for ty := 0; ty < YSIZE; ty++ {

				for tz := 0; tz < ZSIZE; tz++ {
					if dist(x, y, z, tx, ty, tz) <= float64(cL*cL) && tk.legal(tx, ty, tx) {
						var newag agent
						newag.startUp()
						newag.x, newag.y, newag.z = tx, ty, tz
						tk.add(newag.x, newag.y, newag.z)
						tk.ags = append(tk.ags, &newag)
						if len(tk.ags) >= tot {
							println(cL)
							return
						}

					}
				}
			}
		}
	}
}

func (tk *took) exe() {
	tk.ar = make([]bool, SIZE)

	//fmt.Println("x,y,z")
	for t := 0; t < TURNS; t++ {
		//if t % 100 == 0 {fmt.Fprintln(os.Stderr, "t", t)}
		for ags := 0; ags < AGENTSPT; ags++ {
			if t > 0 && !SPAWN_NEW {
				break
			}
			if tk.count >= SIZE {
				panic("disco")
			}
			var x, y, z int
			for {
				//		x = randit(CIRC, tk)
				//               y = randit(CIRC, tk)
				x = XSIZE/2 - CIRC + randit(2*CIRC, tk)
				y = XSIZE/2 - CIRC + randit(2*CIRC, tk)
				x = 150
				y = 150
				x = randit(XSIZE, tk)
				y = randit(YSIZE, tk)
				if randit(100, tk) < tk.rangen {
					z = randit(ZSIZE, tk)
					//										z = ZSIZE - 1 - randit(10, tk)
				} else {
					z = ZSIZE - 1
				}

				if tk.legal(x, y, z) {
					break
				}
			}
			var newag agent
			newag.startUp()
			newag.aid = t*AGENTSPT + ags
			newag.x, newag.y, newag.z = x, y, z
			if newag.z == 0 {
				newag.dead = true
			}
			tk.add(newag.x, newag.y, newag.z)
			tk.ags = append(tk.ags, &newag)
		}
		for _, a := range tk.ags {
			a.iter(tk, t)
			_ = a
		}
	}
	tk.statz()

}

const TH = 1

func gorange(b, e int, tk took, ch chan float64) {
	score := 0.0
	for i := b; i < e; i++ {
		score += tk.ags[i].adist(tk)
	}
	ch <- score
}

type rs struct {
	v    float64
	mad  float64
	dead int
	top  int
	ind  int
}

func all5(prin bool) rs {
	var rt rs

	var rss [6]rs
	mc := make(chan rs)
	for i := 1; i <= 6; i++ {
		go gotook(i, mc)
	}
	for i := 1; i <= 6; i++ {
		rs := <-mc
		rss[rs.ind-1] = rs
		rt.v += rs.v
		rt.mad += rs.mad
		rt.dead += rs.dead
		rt.top += rs.top
	}

	if prin {
		for i := 1; i <= 6; i++ {
			fmt.Fprintln(os.Stderr, i, rss[i-1].v, rss[i-1].mad, rss[i-1].dead, rss[i-1].top)
		}
	}

	return rt
}
func randinit() {
	rand.Seed(SEED)
	xs = SEED
	randi(1)
	randi(1)
}
func hasAcc(a int) bool {
	return a < 6
}

func gotook(a int, ch chan rs) {
	t := took{}
	t.ind = a
	t.startUp()
	t.envar = eas[a-1]
	if hasAcc(a) {
		t.accc = accs[a-1]
	}

	t.exe()
	if a == 6 {
		t.prn()
		//                t.trackp()
	}
	sc, mad := 0.0, 0.0
	if hasAcc(a) {
		sc, mad = t.score()
	}
	dead, top := t.statz()
	ch <- rs{v: sc, ind: a, mad: mad, dead: dead, top: top}
}

var eas [6][]uint8
var accs [5][][3]uint16

func loadFiles() {
	for i := 1; i <= 6; i++ {
		eas[i-1] = loadEA(i)
	}
	for i := 1; i <= 5; i++ {
		accs[i-1] = loadAc(i)
	}

}

const MUT = 30

func mut() {
	Gprefs = GprefsA
	for dz := 0; dz < 2; dz++ {
		cross := randi(MUT)
		diag := randi(MUT)

		Gprefs[dz][1][0] += cross
		Gprefs[dz][1][2] += cross
		Gprefs[dz][0][1] += cross
		Gprefs[dz][2][1] += cross
		Gprefs[dz][0][0] += diag
		Gprefs[dz][0][2] += diag
		Gprefs[dz][2][0] += diag
		Gprefs[dz][2][2] += diag
		if dz == 1 {
			Gprefs[dz][1][1] += randi(MUT)
		}
	}
}

func globSet() string {
	gt := took{}
	gt.startUp()
	return gt.disp()
}
func oneOff() {

}

func main() {

	randinit()

	loadFiles()
	Gprefs = GprefsA
	fmt.Fprintln(os.Stderr, globSet())

	stat := all5(true)
	fmt.Fprintln(os.Stderr, stat.v, stat.mad, stat.dead, stat.top)
	os.Exit(0)

}
