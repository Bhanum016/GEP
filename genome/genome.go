// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Package genome provides the basis for a single GEP genome.
// A genome consists of one or more genes.
package genome

import (
	"log"
	"math/rand"
	"strings"

	"github.com/gmlewis/gep/functions"
	mn "github.com/gmlewis/gep/functions/math_nodes"
	"github.com/gmlewis/gep/gene"
)

// Genome contains the genes that make up the genome.
// It also provides the linking function and the score that results
// from evaluating the genome against the fitness function.
type Genome struct {
	Genes    []*gene.Gene
	LinkFunc string
	Score    float64

	SymbolMap map[string]int // do not use directly.  Use SymbolCount() instead.
}

// New creates a new genome from the given genes and linking function.
func New(genes []*gene.Gene, linkFunc string) *Genome {
	return &Genome{Genes: genes, LinkFunc: linkFunc}
}

func merge(dst *map[string]int, src map[string]int) {
	for k, v := range src {
		(*dst)[k] += v
	}
}

// SymbolCount returns the count of the number of times the symbol
// is actually used in the Genome.
// Note that this count is typically different from the number
// of times the symbol appears in the Karva expression.  This can be
// a handy metric to assist in the fitness evaluation of a Genome.
// Note also that this currently only works for Math expressions.
// Hopefully this restriction will be lifted in the future.
// A workaround for using it with other types is to evaluate the
// Genome, and then g.symbolCount will already be populated.
func (g *Genome) SymbolCount(sym string) int {
	if g.SymbolMap == nil {
		g.SymbolMap = make(map[string]int)
		g.SymbolMap[g.LinkFunc] = len(g.Genes) - 1
		for i := 0; i < len(g.Genes); i++ {
			g.Genes[i].SymbolCount(sym) // force evaluation
			m := g.Genes[i].SymbolMap
			merge(&(g.SymbolMap), m)
		}
	}
	return g.SymbolMap[sym]
}

// EvalBool evaluates the genome as a boolean expression and returns the result.
// in represents the boolean inputs available to the genome.
// fm is the map of available boolean functions to the genome.
func (g *Genome) EvalBool(in []bool, fm functions.FuncMap) bool {
	lf, ok := fm[g.LinkFunc]
	if !ok {
		log.Printf("Unable to find linking function: %v", g.LinkFunc)
		return false
	}
	result := g.Genes[0].EvalBool(in, fm)
	for i := 1; i < len(g.Genes); i++ {
		result = lf.BoolFunction(result, g.Genes[i].EvalBool(in, fm), false, false)
	}
	return result
}

// EvalMath evaluates the genome as a floating-point expression and returns the result.
// in represents the float64 inputs available to the genome.
func (g *Genome) EvalMath(in []float64) float64 {
	lf, ok := mn.Math[g.LinkFunc]
	if !ok {
		log.Printf("Unable to find linking function: %v", g.LinkFunc)
		return 0.0
	}
	result := g.Genes[0].EvalMath(in)
	for i := 1; i < len(g.Genes); i++ {
		result = lf.Float64Function(result, g.Genes[i].EvalMath(in), 0.0, 0.0)
	}
	return result
}

// String returns the Karva representation of the genome.
func (g Genome) String() string {
	result := []string{}
	for _, v := range g.Genes {
		result = append(result, v.String())
	}
	return strings.Join(result, "|"+g.LinkFunc+"|")
}

// Mutate mutates a genome by performing numMutations random symbol exchanges within the genome.
func (g *Genome) Mutate(numMutations int) {
	for i := 0; i < numMutations; i++ {
		n := rand.Intn(len(g.Genes))
		// fmt.Printf("\nMutating gene #%v, before:\n%v\n", n, g.Genes[n])
		g.Genes[n].Mutate()
		// fmt.Printf("after:\n%v\n", g.Genes[n])
	}
}

// Dup duplicates the genome into the provided destination genome.
func (g *Genome) Dup() *Genome {
	if g == nil {
		log.Printf("denome.Dup error: src and dst must be non-nil")
		return nil
	}
	dst := &Genome{
		Genes:    make([]*gene.Gene, len(g.Genes)),
		LinkFunc: g.LinkFunc,
		Score:    g.Score,
	}
	for i := range g.Genes {
		dst.Genes[i] = g.Genes[i].Dup()
	}
	return dst
}

// ScoringFunc is the function that is used to evaluate the fitness of the model.
// Typically, a return value of 0 means that the function is nowhere close to being
// a valid solution and a return value of 1000 (or higher) means a perfect solution.
type ScoringFunc func(g *Genome) float64

// Evaluate scores a genome and sends the result to a channel.
func (g *Genome) Evaluate(sf ScoringFunc, c chan<- *Genome) {
	if sf == nil {
		log.Fatalf("genome.Evaluate: ScoringFunc must not be nil")
	}
	g.Score = sf(g)
	c <- g
}
