package packcalc

import (
	"container/heap"
	"errors"
	"fmt"
	"sort"

	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/models"
)

var (
	ErrInvalidQuantity  = errors.New("invalid quantity")
	ErrNoPackSizes      = errors.New("no pack sizes")
	ErrInvalidPackSizes = errors.New("invalid pack sizes")
)

type Calculator interface {
	// Calculate returns a pack allocation that fulfills quantity using whole packs, minimizing:
	// 1) total items shipped (i.e. sum of packs) and then
	// 2) number of packs.
	//
	// The returned list is sorted by Size descending and contains only allocations with Count > 0.
	Calculate(quantity int, packSizes []models.PackSize) ([]models.PackAllocation, error)
}

type defaultCalculator struct{}

var calculator Calculator = defaultCalculator{}

func CalculatorImpl() Calculator {
	return calculator
}

func SetCalculator(c Calculator) {
	if c == nil {
		panic("Calculator must not be nil")
	}
	calculator = c
}

func Calculate(quantity int, packSizes []models.PackSize) ([]models.PackAllocation, error) {
	return calculator.Calculate(quantity, packSizes)
}

func (defaultCalculator) Calculate(quantity int, packSizes []models.PackSize) ([]models.PackAllocation, error) {
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	sizes := make([]int, 0, len(packSizes))
	for _, p := range packSizes {
		sizes = append(sizes, p.Size)
	}

	sizes, err := normalizePackSizes(sizes)
	if err != nil {
		return nil, err
	}
	if len(sizes) == 0 {
		return nil, ErrNoPackSizes
	}

	minSum, err := minimalShippedAtLeast(quantity, sizes)
	if err != nil {
		return nil, err
	}

	counts, err := minPacksForExactSum(minSum, sizes)
	if err != nil {
		return nil, err
	}

	out := make([]models.PackAllocation, 0, len(sizes))
	for i := len(sizes) - 1; i >= 0; i-- { // descending
		s := sizes[i]
		if counts[s] > 0 {
			out = append(out, models.PackAllocation{Size: s, Count: counts[s]})
		}
	}
	return out, nil
}

func normalizePackSizes(in []int) ([]int, error) {
	if len(in) == 0 {
		return nil, nil
	}
	seen := make(map[int]struct{}, len(in))
	out := make([]int, 0, len(in))
	for _, s := range in {
		if s <= 0 {
			return nil, ErrInvalidPackSizes
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Ints(out)
	return out, nil
}

// minimalShippedAtLeast finds the minimal achievable shipped sum >= quantity.
// It uses Dijkstra over residues modulo the smallest pack size to find the minimal base sum
// for each residue, then lifts each residue by adding the smallest pack size as needed.
func minimalShippedAtLeast(quantity int, sizes []int) (int, error) {
	m := sizes[0]
	const inf = int(^uint(0) >> 1)

	dist := make([]int, m)
	prevRes := make([]int, m)
	prevSize := make([]int, m)
	for i := 0; i < m; i++ {
		dist[i] = inf
		prevRes[i] = -1
		prevSize[i] = 0
	}
	dist[0] = 0

	pq := &resPQ{}
	heap.Init(pq)
	heap.Push(pq, resNode{res: 0, sum: 0})

	for pq.Len() > 0 {
		cur := heap.Pop(pq).(resNode)
		if cur.sum != dist[cur.res] {
			continue
		}
		for _, s := range sizes {
			nr := (cur.res + s) % m
			ns := cur.sum + s
			if ns < dist[nr] {
				dist[nr] = ns
				prevRes[nr] = cur.res
				prevSize[nr] = s
				heap.Push(pq, resNode{res: nr, sum: ns})
			}
		}
	}

	best := inf
	for r := 0; r < m; r++ {
		if dist[r] == inf {
			continue
		}
		cand := dist[r]
		if cand < quantity {
			need := quantity - cand
			k := (need + m - 1) / m
			cand = cand + k*m
		}
		if cand < best {
			best = cand
		}
	}
	if best == inf {
		return 0, fmt.Errorf("no solution")
	}
	return best, nil
}

// minPacksForExactSum computes the minimum number of packs needed to reach exactSum.
// It returns a map[size]count allocation.
func minPacksForExactSum(exactSum int, sizes []int) (map[int]int, error) {
	const inf = int(^uint(0) >> 1)

	dp := make([]int, exactSum+1)
	prevIdx := make([]int, exactSum+1)
	prevSize := make([]int, exactSum+1)
	for i := 1; i <= exactSum; i++ {
		dp[i] = inf
		prevIdx[i] = -1
		prevSize[i] = 0
	}
	dp[0] = 0
	prevIdx[0] = -1

	// Iterate sizes descending so ties tend to prefer larger packs (more stable output).
	sizesDesc := append([]int(nil), sizes...)
	sort.Sort(sort.Reverse(sort.IntSlice(sizesDesc)))

	for i := 0; i <= exactSum; i++ {
		if dp[i] == inf {
			continue
		}
		for _, s := range sizesDesc {
			j := i + s
			if j > exactSum {
				continue
			}
			cand := dp[i] + 1
			if cand < dp[j] || (cand == dp[j] && s > prevSize[j]) {
				dp[j] = cand
				prevIdx[j] = i
				prevSize[j] = s
			}
		}
	}

	if dp[exactSum] == inf {
		return nil, fmt.Errorf("no exact solution for %d", exactSum)
	}

	out := make(map[int]int, len(sizes))
	for cur := exactSum; cur > 0; {
		s := prevSize[cur]
		if s == 0 || prevIdx[cur] < 0 {
			return nil, fmt.Errorf("failed to reconstruct solution for %d", exactSum)
		}
		out[s]++
		cur = prevIdx[cur]
	}
	return out, nil
}

type resNode struct {
	res int
	sum int
}

type resPQ []resNode

func (p resPQ) Len() int           { return len(p) }
func (p resPQ) Less(i, j int) bool { return p[i].sum < p[j].sum }
func (p resPQ) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p *resPQ) Push(x any)        { *p = append(*p, x.(resNode)) }
func (p *resPQ) Pop() any          { old := *p; n := len(old); x := old[n-1]; *p = old[:n-1]; return x }
