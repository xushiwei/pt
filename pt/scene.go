package pt

import (
	"math"
	"math/rand"
)

type Scene struct {
	shapes    []Shape
	lights    []Shape
	shapeTree *Tree
	lightTree *Tree
}

func (s *Scene) Compile() {
	s.shapeTree = NewTree(s.shapes)
	s.lightTree = NewTree(s.lights)
}

func (s *Scene) AddShape(shape Shape) {
	s.shapes = append(s.shapes, shape)
}

func (s *Scene) AddLight(shape Shape) {
	s.lights = append(s.lights, shape)
}

func (s *Scene) IntersectShapes(r Ray) (Hit, bool) {
	return s.shapeTree.Intersect(r)
}

func (s *Scene) IntersectLights(r Ray) (Hit, bool) {
	hit1, ok1 := s.lightTree.Intersect(r)
	if ok1 {
		// TODO: clean this up
		hit2, ok2 := s.shapeTree.Intersect(r)
		if ok2 {
			ok1 = hit1.T < hit2.T
		}
	}
	return hit1, ok1
}

func (s *Scene) Shadow(r Ray, max float64) bool {
	t := s.shapeTree.Shadow(r)
	return t < max
}

func (s *Scene) DirectLight(i, n Ray, rnd *rand.Rand) Color {
	color := Color{}
	for _, light := range s.lights {
		p := light.RandomPoint(rnd)
		d := p.Sub(n.Origin)
		lr := Ray{n.Origin, d.Normalize()}
		if s.Shadow(lr, d.Length()) {
			continue
		}
		diffuse := math.Max(0, lr.Direction.Dot(n.Direction))
		color = color.Add(light.Color(p).Mul(diffuse))
	}
	return color.Div(float64(len(s.lights)))
}

func (s *Scene) RecursiveSample(r Ray, reflected bool, depth int, rnd *rand.Rand) Color {
	if depth < 0 {
		return Color{}
	}
	if reflected {
		hit, ok := s.IntersectLights(r)
		if ok {
			return hit.Shape.Color(hit.Ray.Origin)
		}
	}
	hit, ok := s.IntersectShapes(r)
	if !ok {
		return Color{}
	}
	shape := hit.Shape
	color := shape.Color(hit.Ray.Origin)
	material := shape.Material(hit.Ray.Origin)
	p, u, v := rnd.Float64(), rnd.Float64(), rnd.Float64()
	ray, reflected := hit.Ray.Bounce(r, material, p, u, v)
	indirect := s.RecursiveSample(ray, reflected, depth-1, rnd)
	if reflected {
		if material.Tint > 0 {
			a := color.MulColor(indirect.Mul(material.Tint))
			b := indirect.Mul(1 - material.Tint)
			return a.Add(b)
		} else {
			return indirect
		}
	} else {
		direct := s.DirectLight(r, hit.Ray, rnd)
		return color.MulColor(direct.Add(indirect))
	}
}

func (s *Scene) Sample(r Ray, samples, depth int, rnd *rand.Rand) Color {
	if depth < 0 {
		return Color{}
	}
	hit, ok := s.IntersectShapes(r)
	if !ok {
		return Color{}
	}
	shape := hit.Shape
	color := shape.Color(hit.Ray.Origin)
	material := shape.Material(hit.Ray.Origin)
	result := Color{}
	n := int(math.Sqrt(float64(samples)))
	for u := 0; u < n; u++ {
		for v := 0; v < n; v++ {
			p := rnd.Float64()
			fu := (float64(u) + rnd.Float64()) * (1 / float64(n))
			fv := (float64(v) + rnd.Float64()) * (1 / float64(n))
			ray, reflected := hit.Ray.Bounce(r, material, p, fu, fv)
			indirect := s.RecursiveSample(ray, reflected, depth-1, rnd)
			if reflected {
				if material.Tint > 0 {
					a := color.MulColor(indirect.Mul(material.Tint))
					b := indirect.Mul(1 - material.Tint)
					result = result.Add(a.Add(b))
				} else {
					result = result.Add(indirect)
				}
			} else {
				direct := s.DirectLight(r, hit.Ray, rnd)
				result = result.Add(color.MulColor(direct.Add(indirect)))
			}
		}
	}
	return result.Div(float64(n * n))
}