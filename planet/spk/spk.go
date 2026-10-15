package spk

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/starainrt/astro/internal/ephem"
)

const (
	dafRecordBytes  = 1024
	dafFileHeadSize = 1024
	dafND           = 2
	dafNI           = 6
	dafSummaryWords = dafND + (dafNI+1)/2
	dafSummaryBytes = dafSummaryWords * 8
	dafHeaderWords  = 3
	dafWordBytes    = 8
	spkJ2000        = 2451545.0
	spkSecondsDay   = 86400.0
	spkCacheRecords = 16
)

var (
	// ErrInvalidFile 无效 SPK 文件 / invalid SPK file.
	ErrInvalidFile = errors.New("spk: invalid DAF/SPK file")
	// ErrUnsupportedFrame 不支持的参考框架 / unsupported reference frame.
	ErrUnsupportedFrame = errors.New("spk: unsupported reference frame")
	// ErrUnsupportedType 不支持的段类型 / unsupported segment type.
	ErrUnsupportedType = errors.New("spk: unsupported SPK data type")
	// ErrClosed 文件已关闭 / file is closed.
	ErrClosed     = errors.New("spk: file is closed")
	errInvalidDAF = ErrInvalidFile
	errClosed     = ErrClosed
)

// Body NAIF 天体编号 / NAIF body identifier.
type Body = ephem.Body

const (
	// SSB 太阳系质心 / solar-system barycenter.
	SSB = ephem.SSB
	// Mercury 水星 / Mercury.
	Mercury = ephem.Mercury
	// Venus 金星 / Venus.
	Venus = ephem.Venus
	// EMB 地月质心 / Earth-Moon barycenter.
	EMB = ephem.EMB
	// Mars 火星 / Mars.
	Mars = ephem.Mars
	// Jupiter 木星 / Jupiter.
	Jupiter = ephem.Jupiter
	// Saturn 土星 / Saturn.
	Saturn = ephem.Saturn
	// Uranus 天王星 / Uranus.
	Uranus = ephem.Uranus
	// Neptune 海王星 / Neptune.
	Neptune = ephem.Neptune
	// Sun 太阳 / the Sun.
	Sun = ephem.Sun
	// Moon 月球 / the Moon.
	Moon = ephem.Moon
	// Earth 地球 / Earth.
	Earth = ephem.Earth
)

// Segment 描述一个支持的 SPK 段 / describes a supported SPK segment.
type Segment struct {
	Target Body
	Center Body
	Frame  int
	Type   int
	Start  float64
	End    float64
}

type summarySegment struct {
	target         Body
	center         Body
	frame          int
	dataType       int
	startET        float64
	endET          float64
	startWord      int64
	endWord        int64
	initET         float64
	interval       float64
	recordSize     int
	records        int
	maxDim         int
	distanceScale  float64
	timeScale      float64
	initJD         float64
	initFR         float64
	epochStartWord int64
	directoryCount int
	cacheMu        sync.Mutex
	cache          map[int][]float64
	cacheOrder     []int
	owner          *File
}

// File 只读 SPK 文件 / read-only SPK file.
type File struct {
	mu       sync.RWMutex
	file     *os.File
	order    binary.ByteOrder
	segments []*summarySegment
	known    map[Body]bool
	closed   bool
	size     int64
}

// Open 打开并解析 SPK 文件 / opens and indexes an SPK file.
func Open(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	spk := &File{file: f, known: make(map[Body]bool)}
	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	spk.size = stat.Size()
	if err := spk.index(); err != nil {
		_ = f.Close()
		return nil, err
	}
	return spk, nil
}

// Load Open 的别名 / alias for Open.
func Load(path string) (*File, error) { return Open(path) }

// Close 关闭文件 / closes the file.
func (s *File) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return s.file.Close()
}

// Segments 返回支持段的快照 / returns a snapshot of supported segments.
func (s *File) Segments() []Segment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Segment, 0, len(s.segments))
	for _, segment := range s.segments {
		result = append(result, Segment{
			Target: segment.target,
			Center: segment.center,
			Frame:  segment.frame,
			Type:   segment.dataType,
			Start:  spkJ2000 + segment.startET/spkSecondsDay,
			End:    spkJ2000 + segment.endET/spkSecondsDay,
		})
	}
	return result
}

// Coverage 返回天体覆盖的 TDB 儒略日范围 / returns the body's TDB Julian-date coverage.
func (s *File) Coverage(target Body) (start, end float64, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return 0, 0, false
	}
	for _, segment := range s.segments {
		if segment.target != target || segment.frame != 1 {
			continue
		}
		segmentStart := spkJ2000 + segment.startET/spkSecondsDay
		segmentEnd := spkJ2000 + segment.endET/spkSecondsDay
		if !ok || segmentStart < start {
			start = segmentStart
		}
		if !ok || segmentEnd > end {
			end = segmentEnd
		}
		ok = true
	}
	return start, end, ok
}

// Bodies 返回文件中出现的天体编号 / returns body identifiers present in the file.
func (s *File) Bodies() []Body {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return nil
	}
	result := make([]Body, 0, len(s.known))
	for body := range s.known {
		result = append(result, body)
	}
	return result
}

// State 返回目标相对中心的 ICRF 状态，速度为 km/日 / returns an ICRF state, with velocity in km/day.
func (s *File) State(target, center Body, jdTDB float64) (pos, vel [3]float64, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return pos, vel, errClosed
	}
	if math.IsNaN(jdTDB) || math.IsInf(jdTDB, 0) {
		return pos, vel, ephem.ErrOutOfRange
	}
	if target == center {
		return pos, vel, nil
	}
	return s.composeState(target, center, jdTDB)
}

func (s *File) index() error {
	header := make([]byte, dafFileHeadSize)
	if _, err := s.file.ReadAt(header, 0); err != nil {
		return err
	}
	if string(header[:8]) != "DAF/SPK " {
		return errInvalidDAF
	}
	order, err := byteOrderFor(string(header[88:96]))
	if err != nil {
		return err
	}
	s.order = order
	if int32Value(s.order, header, 8) != dafND || int32Value(s.order, header, 12) != dafNI {
		return fmt.Errorf("%w: ND/NI must be 2/6", errInvalidDAF)
	}
	first := int(int32Value(s.order, header, 76))
	last := int(int32Value(s.order, header, 80))
	if first < 2 || last < 2 || int64(first)*dafRecordBytes > s.size || int64(last)*dafRecordBytes > s.size {
		return fmt.Errorf("%w: invalid summary record range", errInvalidDAF)
	}
	seen := make(map[int]bool)
	previous := 0
	supported := 0
	for record := first; record != 0; {
		if seen[record] {
			return fmt.Errorf("%w: cyclic summary record chain", errInvalidDAF)
		}
		seen[record] = true
		data, err := s.readRecord(record)
		if err != nil {
			return err
		}
		nextValue := float64Value(s.order, data[0:8])
		previousValue := float64Value(s.order, data[8:16])
		countValue := float64Value(s.order, data[16:24])
		if !validInteger(nextValue, 0, s.size/dafRecordBytes) || previousValue != float64(previous) || !validInteger(countValue, 0, (dafRecordBytes-24)/dafSummaryBytes) {
			return fmt.Errorf("%w: invalid summary controls", errInvalidDAF)
		}
		next, count := int(nextValue), int(countValue)
		if next == 0 && record != last {
			return fmt.Errorf("%w: final summary disagrees with backward pointer", errInvalidDAF)
		}
		for index := 0; index < count; index++ {
			base := dafHeaderWords*8 + index*dafSummaryBytes
			segment, err := s.parseSummary(data[base : base+dafSummaryBytes])
			if err != nil {
				return err
			}
			if segment.frame == 1 && (segment.dataType == 2 || segment.dataType == 3 || segment.dataType == 20 || segment.dataType == 21) {
				if err := segment.loadMeta(s); err != nil {
					return err
				}
				supported++
			} else {
				continue
			}
			s.segments = append(s.segments, segment)
			s.known[segment.target] = true
			s.known[segment.center] = true
		}
		previous, record = record, next
	}
	if supported == 0 {
		return fmt.Errorf("%w: no supported J2000 segments", errInvalidDAF)
	}
	return nil
}

func (s *File) parseSummary(data []byte) (*summarySegment, error) {
	if len(data) != dafSummaryBytes {
		return nil, errInvalidDAF
	}
	segment := &summarySegment{
		startET:   float64Value(s.order, data[0:8]),
		endET:     float64Value(s.order, data[8:16]),
		target:    Body(int32Value(s.order, data, 16)),
		center:    Body(int32Value(s.order, data, 20)),
		frame:     int(int32Value(s.order, data, 24)),
		dataType:  int(int32Value(s.order, data, 28)),
		startWord: int64(int32Value(s.order, data, 32)),
		endWord:   int64(int32Value(s.order, data, 36)),
		cache:     make(map[int][]float64),
		owner:     s,
	}
	if !finite(segment.startET) || !finite(segment.endET) || segment.endET < segment.startET || segment.startWord < 1 || segment.endWord < segment.startWord || segment.endWord > s.size/dafWordBytes {
		return nil, fmt.Errorf("%w: invalid segment summary", errInvalidDAF)
	}
	return segment, nil
}

func (segment *summarySegment) loadMeta(s *File) error {
	if segment.endWord-segment.startWord+1 < 4 {
		return fmt.Errorf("%w: segment trailer is missing", errInvalidDAF)
	}
	if segment.dataType == 20 {
		trailer, err := s.readWords(segment.endWord-6, 7)
		if err != nil {
			return err
		}
		if !finitePositive(trailer[0]) || !finitePositive(trailer[1]) || !finite(trailer[2]) || math.Abs(trailer[3]) >= 1 || !finitePositive(trailer[4]) || !validInteger(trailer[5], 6, segment.endWord-segment.startWord) || !validInteger(trailer[6], 1, segment.endWord-segment.startWord) {
			return fmt.Errorf("%w: invalid type 20 trailer", errInvalidDAF)
		}
		segment.distanceScale = trailer[0]
		segment.timeScale = trailer[1]
		segment.initJD = trailer[2]
		segment.initFR = trailer[3]
		segment.interval = trailer[4]
		segment.recordSize = int(trailer[5])
		segment.records = int(trailer[6])
		wordCount := segment.endWord - segment.startWord - 6
		if int64(segment.recordSize)*int64(segment.records) != wordCount || segment.recordSize < 6 || segment.recordSize%3 != 0 {
			return fmt.Errorf("%w: invalid type 20 record size", errInvalidDAF)
		}
		return nil
	}
	if segment.dataType == 21 {
		trailer, err := s.readWords(segment.endWord-1, 2)
		if err != nil {
			return err
		}
		if !validInteger(trailer[0], 1, 25) || !validInteger(trailer[1], 1, segment.endWord-segment.startWord) {
			return fmt.Errorf("%w: invalid type 21 trailer", errInvalidDAF)
		}
		segment.maxDim = int(trailer[0])
		segment.recordSize = 4*segment.maxDim + 11
		segment.records = int(trailer[1])
		segment.epochStartWord = segment.startWord + int64(segment.recordSize*segment.records)
		segment.directoryCount = segment.records / 100
		wordCount := segment.endWord - segment.startWord + 1
		want := int64(segment.recordSize*segment.records + segment.records + segment.directoryCount + 2)
		if want != wordCount {
			return fmt.Errorf("%w: invalid type 21 segment size", errInvalidDAF)
		}
		return nil
	}
	trailerWords := segment.endWord - 3
	trailer, err := s.readWords(trailerWords, 4)
	if err != nil {
		return err
	}
	segment.initET = trailer[0]
	segment.interval = trailer[1]
	wordCount := segment.endWord - segment.startWord - 3
	if !finite(segment.initET) || !finitePositive(segment.interval) || !validInteger(trailer[2], 5, wordCount) || !validInteger(trailer[3], 1, wordCount) {
		return fmt.Errorf("%w: invalid type %d trailer", errInvalidDAF, segment.dataType)
	}
	segment.recordSize = int(trailer[2])
	segment.records = int(trailer[3])
	if int64(segment.recordSize)*int64(segment.records) != wordCount {
		return fmt.Errorf("%w: invalid type %d trailer", errInvalidDAF, segment.dataType)
	}
	components := 3
	if segment.dataType == 3 {
		components = 6
	}
	if (segment.recordSize-2)%components != 0 || (segment.recordSize-2)/components < 1 {
		return fmt.Errorf("%w: invalid type %d record size %d", errInvalidDAF, segment.dataType, segment.recordSize)
	}
	return nil
}

func chosenSegment(segment *summarySegment) (*summarySegment, error) {
	if segment == nil {
		return nil, ephem.ErrOutOfRange
	}
	if segment.frame != 1 {
		return nil, ErrUnsupportedFrame
	}
	if segment.dataType != 2 && segment.dataType != 3 && segment.dataType != 20 && segment.dataType != 21 {
		return nil, ErrUnsupportedType
	}
	return segment, nil
}

type stateEdge struct {
	to      Body
	segment *summarySegment
	reverse bool
}

func (s *File) composeState(target, center Body, jdTDB float64) (pos, vel [3]float64, err error) {
	if !s.known[target] || !s.known[center] {
		return pos, vel, ephem.ErrUnsupportedBody
	}
	// 一次倒序扫描求出每个目标在 jdTDB 处覆盖它的最后一段，后加载的段优先。
	chosen := make(map[Body]*summarySegment)
	for index := len(s.segments) - 1; index >= 0; index-- {
		segment := s.segments[index]
		if _, ok := chosen[segment.target]; ok {
			continue
		}
		if jdTDB >= spkJ2000+segment.startET/spkSecondsDay && jdTDB <= spkJ2000+segment.endET/spkSecondsDay {
			chosen[segment.target] = segment
		}
	}
	graph := make(map[Body][]stateEdge)
	selected := make(map[Body]bool)
	for index := len(s.segments) - 1; index >= 0; index-- {
		segment := s.segments[index]
		if selected[segment.target] {
			continue
		}
		current, selectErr := chosenSegment(chosen[segment.target])
		if selectErr != nil {
			// 覆盖缺口留到 BFS 结束再报：别的路径仍可能连通。
			if selectErr != ephem.ErrOutOfRange && (segment.target == target || segment.target == center) {
				return pos, vel, selectErr
			}
			continue
		}
		segment = current
		selected[segment.target] = true
		graph[segment.target] = append(graph[segment.target], stateEdge{to: segment.center, segment: segment})
		graph[segment.center] = append(graph[segment.center], stateEdge{to: segment.target, segment: segment, reverse: true})
	}
	type node struct {
		body Body
		pos  [3]float64
		vel  [3]float64
	}
	queue := []node{{body: target}}
	visited := map[Body]bool{target: true}
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		if current.body == center {
			return current.pos, current.vel, nil
		}
		for _, edge := range graph[current.body] {
			if visited[edge.to] {
				continue
			}
			edgePos, edgeVel, edgeErr := edge.segment.state(jdTDB)
			if edgeErr != nil {
				return pos, vel, edgeErr
			}
			if edge.reverse {
				negateVector(&edgePos)
				negateVector(&edgeVel)
			}
			addVector(&edgePos, current.pos)
			addVector(&edgeVel, current.vel)
			visited[edge.to] = true
			queue = append(queue, node{body: edge.to, pos: edgePos, vel: edgeVel})
		}
	}
	return pos, vel, ephem.ErrOutOfRange
}

func (segment *summarySegment) state(jdTDB float64) (pos, vel [3]float64, err error) {
	if segment.dataType == 20 {
		return segment.stateType20(jdTDB)
	}
	if segment.dataType == 21 {
		return segment.stateType21(jdTDB)
	}
	et := (jdTDB - spkJ2000) * spkSecondsDay
	index := int(math.Floor((et - segment.initET) / segment.interval))
	if index < 0 {
		index = 0
	}
	if index >= segment.records {
		index = segment.records - 1
	}
	if jdTDB < spkJ2000+segment.startET/spkSecondsDay || jdTDB > spkJ2000+segment.endET/spkSecondsDay {
		return pos, vel, ephem.ErrOutOfRange
	}
	record, err := segment.readRecord(index)
	if err != nil {
		return pos, vel, err
	}
	if !finite(record[0]) || !finitePositive(record[1]) || !finiteAll(record[2:]) {
		return pos, vel, errInvalidDAF
	}
	x := (et - record[0]) / record[1]
	tolerance := math.Abs(math.Nextafter(jdTDB, math.Inf(1))-jdTDB) * spkSecondsDay / record[1]
	if x < -1-tolerance || x > 1+tolerance {
		return pos, vel, ephem.ErrOutOfRange
	}
	x = math.Max(-1, math.Min(1, x))
	components := 3
	if segment.dataType == 3 {
		components = 6
	}
	coefficientCount := (segment.recordSize - 2) / components
	for component := 0; component < 3; component++ {
		value, derivative := chebyshev(record[2+component*coefficientCount:2+(component+1)*coefficientCount], x)
		pos[component] = value
		if segment.dataType == 2 {
			vel[component] = derivative / record[1] * spkSecondsDay
		} else {
			velocityStart := 2 + (component+3)*coefficientCount
			vel[component], _ = chebyshev(record[velocityStart:velocityStart+coefficientCount], x)
			vel[component] *= spkSecondsDay
		}
	}
	return pos, vel, nil
}

func (segment *summarySegment) stateType20(jdTDB float64) (pos, vel [3]float64, err error) {
	if jdTDB < spkJ2000+segment.startET/spkSecondsDay || jdTDB > spkJ2000+segment.endET/spkSecondsDay {
		return pos, vel, ephem.ErrOutOfRange
	}
	et := (jdTDB - spkJ2000) * spkSecondsDay
	init := (segment.initJD - spkJ2000 + segment.initFR) * spkSecondsDay
	interval := segment.interval * spkSecondsDay
	if !finitePositive(interval) {
		return pos, vel, errInvalidDAF
	}
	index := int(math.Floor((et - init) / interval))
	if index < 0 {
		index = 0
	}
	if index >= segment.records {
		index = segment.records - 1
	}
	record, err := segment.readRecord(index)
	if err != nil {
		return pos, vel, err
	}
	nterms := segment.recordSize / 3
	coefficientCount := nterms - 1
	if len(record) != segment.recordSize || coefficientCount < 1 || !finiteAll(record) {
		return pos, vel, errInvalidDAF
	}
	recbeg := (segment.initJD - spkJ2000 + float64(index)*segment.interval) * spkSecondsDay
	radius := interval / 2
	mid := recbeg + segment.initFR*spkSecondsDay + radius
	x := (et - mid) / radius
	if !finite(x) || x < -1-1e-12 || x > 1+1e-12 {
		return pos, vel, ephem.ErrOutOfRange
	}
	x = math.Max(-1, math.Min(1, x))
	velocityScale := segment.distanceScale / segment.timeScale
	for component := 0; component < 3; component++ {
		base := component * nterms
		coefficients := record[base : base+coefficientCount]
		value, _ := chebyshev(coefficients, x)
		integral := chebyshevIntegral(coefficients, x)
		pos[component] = record[base+coefficientCount]*segment.distanceScale + radius*integral*velocityScale
		vel[component] = value * velocityScale * spkSecondsDay
	}
	return pos, vel, nil
}

func (segment *summarySegment) stateType21(jdTDB float64) (pos, vel [3]float64, err error) {
	if jdTDB < spkJ2000+segment.startET/spkSecondsDay || jdTDB > spkJ2000+segment.endET/spkSecondsDay {
		return pos, vel, ephem.ErrOutOfRange
	}
	et := (jdTDB - spkJ2000) * spkSecondsDay
	epochs, err := segment.owner.readWords(segment.epochStartWord, segment.records)
	if err != nil {
		return pos, vel, err
	}
	index := 0
	for index < len(epochs) && epochs[index] < et {
		index++
	}
	if index >= len(epochs) {
		index = len(epochs) - 1
	}
	record, err := segment.readRecord(index)
	if err != nil {
		return pos, vel, err
	}
	maxdim := segment.maxDim
	record = append([]float64{float64(maxdim)}, record...)
	pos, vel, err = evaluateType21(et, record)
	for component := range vel {
		vel[component] *= spkSecondsDay
	}
	return pos, vel, err
}

func evaluateType21(et float64, record []float64) (statePos, stateVel [3]float64, err error) {
	if len(record) < 72 {
		return statePos, stateVel, errInvalidDAF
	}
	maxdim := int(math.Round(record[0]))
	if maxdim < 15 || maxdim > 25 || len(record) != 4*maxdim+12 || !finiteAll(record) {
		return statePos, stateVel, errInvalidDAF
	}
	tl := record[1]
	g := record[2 : 2+maxdim]
	refPos := [3]float64{record[maxdim+2], record[maxdim+4], record[maxdim+6]}
	refVel := [3]float64{record[maxdim+3], record[maxdim+5], record[maxdim+7]}
	kqmax1 := int(math.Round(record[4*maxdim+8]))
	kq := [3]int{int(math.Round(record[4*maxdim+9])), int(math.Round(record[4*maxdim+10])), int(math.Round(record[4*maxdim+11]))}
	// 差分表与步长表各只有 MAXDIM 项，权重表只有 KQMAX1+1 项。
	if kqmax1 < 3 || kqmax1 > maxdim+1 {
		return statePos, stateVel, errInvalidDAF
	}
	for _, value := range g[:kqmax1-2] {
		if !finite(value) || value == 0 {
			return statePos, stateVel, errInvalidDAF
		}
	}
	for _, value := range kq {
		if value < 1 || value > kqmax1-1 {
			return statePos, stateVel, errInvalidDAF
		}
	}
	delta := et - tl
	tp := delta
	mq2 := kqmax1 - 2
	ks := kqmax1 - 1
	fc := make([]float64, kqmax1+1)
	wc := make([]float64, kqmax1)
	weights := make([]float64, kqmax1+2)
	for j := 1; j <= mq2; j++ {
		fc[j] = tp / g[j-1]
		wc[j-1] = delta / g[j-1]
		tp = delta + g[j-1]
	}
	for j := 1; j <= kqmax1; j++ {
		weights[j-1] = 1 / float64(j)
	}
	jx := 0
	ks1 := ks - 1
	for ks >= 2 {
		jx++
		for j := 1; j <= jx; j++ {
			weights[j+ks-1] = fc[j]*weights[j+ks1-1] - wc[j-1]*weights[j+ks-1]
		}
		ks = ks1
		ks1--
	}
	for component := 0; component < 3; component++ {
		sum := 0.0
		for j := kq[component]; j >= 1; j-- {
			sum += record[8+(component+1)*maxdim+j-1] * weights[j+ks-1]
		}
		statePos[component] = refPos[component] + delta*(refVel[component]+delta*sum)
	}
	for j := 1; j <= jx; j++ {
		weights[j+ks-1] = fc[j]*weights[j+ks1-1] - wc[j-1]*weights[j+ks-1]
	}
	ks--
	for component := 0; component < 3; component++ {
		sum := 0.0
		for j := kq[component]; j >= 1; j-- {
			sum += record[8+(component+1)*maxdim+j-1] * weights[j+ks-1]
		}
		stateVel[component] = refVel[component] + delta*sum
	}
	return statePos, stateVel, nil
}

func chebyshevIntegral(coefficients []float64, x float64) float64 {
	degree := len(coefficients) - 1
	if degree < 0 {
		return 0
	}
	s2 := 2 * x
	a2 := coefficients[0]
	if len(coefficients) >= 3 {
		a2 -= coefficients[2] * 0.5
	}
	adegree1, adegree2 := 0.0, 0.0
	if degree >= 2 {
		adegree1 = coefficients[degree-1] * 0.5 / float64(degree)
	}
	if degree >= 1 {
		adegree2 = coefficients[degree] * 0.5 / float64(degree+1)
	}
	f2, f1 := 0.0, 0.0
	f0 := a2
	if degree > 0 {
		f0 = adegree2
	}
	z2, z1, z0 := 0.0, 0.0, f0
	w2, w1, w0 := 0.0, 0.0, 0.0
	for index := len(coefficients); index > 1; index-- {
		var coefficient float64
		switch {
		case index == 2:
			coefficient = a2
		case index < len(coefficients):
			coefficient = (coefficients[index-2] - coefficients[index]) * 0.5 / float64(index-1)
		default:
			coefficient = adegree1
		}
		f2, f1 = f1, f0
		f0 = coefficient + (s2*f1 - f2)
		z2, z1 = z1, z0
		z0 = coefficient - z2
		w2, w1 = w1, w0
		w0 = coefficients[index-1] + (s2*w1 - w2)
	}
	c0 := z1
	return c0 + x*f0 - f1
}

func (segment *summarySegment) readRecord(index int) ([]float64, error) {
	if index < 0 || index >= segment.records {
		return nil, ephem.ErrOutOfRange
	}
	segment.cacheMu.Lock()
	if record, ok := segment.cache[index]; ok {
		copyRecord := append([]float64(nil), record...)
		segment.cacheMu.Unlock()
		return copyRecord, nil
	}
	segment.cacheMu.Unlock()
	record, err := segment.owner.readWords(segment.startWord+int64(index*segment.recordSize), segment.recordSize)
	if err != nil {
		return nil, err
	}
	segment.cacheMu.Lock()
	defer segment.cacheMu.Unlock()
	if previous, ok := segment.cache[index]; ok {
		return append([]float64(nil), previous...), nil
	}
	if len(segment.cacheOrder) >= spkCacheRecords {
		delete(segment.cache, segment.cacheOrder[0])
		segment.cacheOrder = segment.cacheOrder[1:]
	}
	segment.cache[index] = record
	segment.cacheOrder = append(segment.cacheOrder, index)
	return append([]float64(nil), record...), nil
}

func chebyshev(coefficients []float64, x float64) (value, derivative float64) {
	t0, t1 := 1.0, x
	d0, d1 := 0.0, 1.0
	for index, coefficient := range coefficients {
		switch index {
		case 0:
			value += coefficient
		case 1:
			value += coefficient * t1
			derivative += coefficient
		default:
			t2 := 2*x*t1 - t0
			d2 := 2*t1 + 2*x*d1 - d0
			value += coefficient * t2
			derivative += coefficient * d2
			t0, t1 = t1, t2
			d0, d1 = d1, d2
		}
	}
	return value, derivative
}

func (s *File) readRecord(record int) ([]byte, error) {
	if record < 1 || int64(record)*dafRecordBytes > s.size {
		return nil, errInvalidDAF
	}
	data := make([]byte, dafRecordBytes)
	if _, err := s.file.ReadAt(data, int64(record-1)*dafRecordBytes); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *File) readWords(startWord int64, count int) ([]float64, error) {
	if startWord < 1 || count < 1 || int64(count) > s.size/dafWordBytes || startWord > s.size/dafWordBytes-int64(count)+1 {
		return nil, errInvalidDAF
	}
	data := make([]byte, count*dafWordBytes)
	if _, err := s.file.ReadAt(data, (startWord-1)*dafWordBytes); err != nil {
		return nil, err
	}
	words := make([]float64, count)
	for index := range words {
		words[index] = float64Value(s.order, data[index*dafWordBytes:(index+1)*dafWordBytes])
	}
	return words, nil
}

func byteOrderFor(format string) (binary.ByteOrder, error) {
	switch format {
	case "LTL-IEEE":
		return binary.LittleEndian, nil
	case "BIG-IEEE":
		return binary.BigEndian, nil
	default:
		return nil, fmt.Errorf("%w: binary format %q", errInvalidDAF, format)
	}
}

func int32Value(order binary.ByteOrder, data []byte, offset int) int32 {
	return int32(order.Uint32(data[offset : offset+4]))
}

func float64Value(order binary.ByteOrder, data []byte) float64 {
	return math.Float64frombits(order.Uint64(data))
}

func finitePositive(value float64) bool {
	return value > 0 && finite(value)
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func finiteAll(values []float64) bool {
	for _, value := range values {
		if !finite(value) {
			return false
		}
	}
	return true
}

func validInteger(value float64, minimum, maximum int64) bool {
	return finite(value) && value >= float64(minimum) && value <= float64(maximum) && value == math.Trunc(value)
}

func addVector(dst *[3]float64, src [3]float64) {
	for index := range dst {
		dst[index] += src[index]
	}
}

func negateVector(value *[3]float64) {
	for index := range value {
		value[index] = -value[index]
	}
}

var _ ephem.Engine = (*File)(nil)
