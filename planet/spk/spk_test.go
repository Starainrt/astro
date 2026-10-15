package spk

import (
	"bufio"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
)

// de441Path 真实 DE441 核文件路径，环境变量 ASTRO_DE441_BSP 优先。
func de441Path() string {
	if path := os.Getenv("ASTRO_DE441_BSP"); path != "" {
		return path
	}
	return "/tmp/de441data/de441_part-2.bsp"
}

type testSPKSegment struct {
	target Body
	center Body
	frame  int
	typ    int
	start  float64
	end    float64
	words  []float64
}

func TestType2StateCoverageAndBoundary(t *testing.T) {
	path := writeTestSPK(t, binary.LittleEndian, "LTL-IEEE", testSPKSegment{
		target: Mars,
		center: Sun,
		frame:  1,
		typ:    2,
		start:  -10,
		end:    10,
		words: type2Words(-10, 10,
			[]float64{-5, 5, 10, 2.5, 0, 0, 20, 0},
			[]float64{5, 5, 20, 2.5, 0, 0, 40, 0},
		),
	})
	spk, err := Open(path)
	if err != nil {
		t.Fatalf("open type21: %+v", err)
	}
	defer spk.Close()
	start, end, ok := spk.Coverage(Mars)
	if !ok || math.Abs(start-(spkJ2000-10/spkSecondsDay)) > 1e-12 || math.Abs(end-(spkJ2000+10/spkSecondsDay)) > 1e-12 {
		t.Fatalf("coverage = %.15f..%.15f/%v", start, end, ok)
	}
	pos, vel, err := spk.State(Mars, Sun, spkJ2000-5/spkSecondsDay)
	if err != nil {
		t.Fatal(err)
	}
	assertVector(t, pos, [3]float64{10, 0, 20})
	assertVector(t, vel, [3]float64{43200, 0, 0})
	pos, _, err = spk.State(Mars, Sun, spkJ2000+10/spkSecondsDay)
	if err != nil {
		t.Fatal(err)
	}
	assertVector(t, pos, [3]float64{22.5, 0, 40})
	_, _, err = spk.State(Mars, Sun, spkJ2000+11/spkSecondsDay)
	if err != ephem.ErrOutOfRange {
		t.Fatalf("out-of-range error = %v", err)
	}
}

func TestType3BigEndianAndStateComposition(t *testing.T) {
	path := writeTestSPK(t, binary.BigEndian, "BIG-IEEE",
		testSPKSegment{
			target: Earth,
			center: EMB,
			frame:  1,
			typ:    3,
			start:  -10,
			end:    10,
			words:  type3Words(0, 10, []float64{1, 2, 3}, []float64{0.1, 0.2, 0.3}),
		},
		testSPKSegment{
			target: EMB,
			center: SSB,
			frame:  1,
			typ:    2,
			start:  -10,
			end:    10,
			words:  type2Words(-10, 20, []float64{0, 10, 10, 20, 30}),
		},
	)
	spk, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer spk.Close()
	pos, vel, err := spk.State(Earth, SSB, spkJ2000)
	if err != nil {
		t.Fatal(err)
	}
	assertVector(t, pos, [3]float64{11, 22, 33})
	assertVector(t, vel, [3]float64{0.1 * spkSecondsDay, 0.2 * spkSecondsDay, 0.3 * spkSecondsDay})
	pos, vel, err = spk.State(SSB, Earth, spkJ2000)
	if err != nil {
		t.Fatal(err)
	}
	assertVector(t, pos, [3]float64{-11, -22, -33})
	assertVector(t, vel, [3]float64{-0.1 * spkSecondsDay, -0.2 * spkSecondsDay, -0.3 * spkSecondsDay})
}

func TestType20State(t *testing.T) {
	intervalDays := 2 / spkSecondsDay
	record := []float64{1, 10, 1, 20, 1, 30}
	path := writeTestSPK(t, binary.LittleEndian, "LTL-IEEE", testSPKSegment{
		target: Mars,
		center: Sun,
		frame:  1,
		typ:    20,
		start:  -1,
		end:    1,
		words:  append(record, []float64{1, 1, spkJ2000, 0, intervalDays, 6, 1}...),
	})
	spk, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer spk.Close()
	pos, vel, err := spk.State(Mars, Sun, spkJ2000+1/spkSecondsDay)
	if err != nil {
		t.Fatal(err)
	}
	assertVector(t, pos, [3]float64{10, 20, 30})
	assertVector(t, vel, [3]float64{spkSecondsDay, spkSecondsDay, spkSecondsDay})
}

func TestType21State(t *testing.T) {
	maxdim := 15
	full := make([]float64, 4*maxdim+12)
	full[0] = 15
	full[1] = 0
	for index := 0; index < maxdim; index++ {
		full[2+index] = 1
	}
	full[maxdim+2] = 10
	full[maxdim+4] = 20
	full[maxdim+6] = 30
	full[maxdim+3] = 1
	full[maxdim+5] = 2
	full[maxdim+7] = 3
	full[8+maxdim] = 1
	full[8+2*maxdim] = 1
	full[8+3*maxdim] = 1
	full[4*maxdim+8] = 3
	full[4*maxdim+9] = 1
	full[4*maxdim+10] = 1
	full[4*maxdim+11] = 1
	words := append(append([]float64(nil), full[1:]...), 2, 15, 1)
	path := writeTestSPK(t, binary.LittleEndian, "LTL-IEEE", testSPKSegment{
		target: Mars,
		center: Sun,
		frame:  1,
		typ:    21,
		start:  0,
		end:    2,
		words:  words,
	})
	spk, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer spk.Close()
	pos, vel, err := spk.State(Mars, Sun, spkJ2000+2/spkSecondsDay)
	if err != nil {
		t.Fatal(err)
	}
	assertVectorTolerance(t, pos, [3]float64{14, 26, 38}, 1e-3)
	assertVectorTolerance(t, vel, [3]float64{3 * spkSecondsDay, 4 * spkSecondsDay, 5 * spkSecondsDay}, 2)
}

func TestLaterSegmentHasPriorityAndClose(t *testing.T) {
	path := writeTestSPK(t, binary.LittleEndian, "LTL-IEEE",
		testSPKSegment{target: Venus, center: Sun, frame: 1, typ: 2, start: -1, end: 1, words: type2Words(-1, 2, []float64{0, 1, 1, 0, 0})},
		testSPKSegment{target: Venus, center: Sun, frame: 1, typ: 2, start: -1, end: 1, words: type2Words(-1, 2, []float64{0, 2, 2, 0, 0})},
	)
	spk, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	pos, _, err := spk.State(Venus, Sun, spkJ2000)
	if err != nil {
		t.Fatal(err)
	}
	assertVector(t, pos, [3]float64{2, 0, 0})
	if err := spk.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := spk.State(Venus, Sun, spkJ2000); err != errClosed {
		t.Fatalf("closed state error = %v", err)
	}
}

func TestOpenRejectsUnsupportedSPK(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.bsp")
	if err := os.WriteFile(path, []byte("not an SPK"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(path); err == nil {
		t.Fatal("invalid SPK was accepted")
	}
}

func type2Words(init, interval float64, records ...[]float64) []float64 {
	if len(records) == 0 || len(records[0]) < 5 || (len(records[0])-2)%3 != 0 {
		panic("invalid type2 test record")
	}
	recordSize := len(records[0])
	words := make([]float64, 0, recordSize*len(records)+4)
	for _, record := range records {
		if len(record) != recordSize {
			panic("type2 record sizes differ")
		}
		words = append(words, record...)
	}
	words = append(words, init, interval, float64(recordSize), float64(len(records)))
	return words
}

func type3Words(mid, radius float64, position, velocity []float64) []float64 {
	words := []float64{mid, radius}
	words = append(words, position...)
	words = append(words, velocity...)
	words = append(words, mid-radius, radius*2, 2+6, 1)
	return words
}

func writeTestSPK(t *testing.T, order binary.ByteOrder, format string, segments ...testSPKSegment) string {
	t.Helper()
	const summaryRecord = 2
	const nameRecord = 3
	const firstDataRecord = 4
	wordsPerRecord := int64(dafRecordBytes / dafWordBytes)
	dataWord := int64(firstDataRecord-1)*wordsPerRecord + 1
	type indexedSegment struct {
		testSPKSegment
		startWord int64
		endWord   int64
	}
	indexed := make([]indexedSegment, len(segments))
	for index, segment := range segments {
		indexed[index] = indexedSegment{testSPKSegment: segment, startWord: dataWord}
		dataWord += int64(len(segment.words))
		indexed[index].endWord = dataWord - 1
	}
	dataRecords := int((dataWord-1+wordsPerRecord-1)/wordsPerRecord) + 1
	fileData := make([]byte, dataRecords*dafRecordBytes)
	copy(fileData[0:8], []byte("DAF/SPK "))
	order.PutUint32(fileData[8:12], uint32(dafND))
	order.PutUint32(fileData[12:16], uint32(dafNI))
	copy(fileData[16:28], []byte("test SPK"))
	order.PutUint32(fileData[76:80], summaryRecord)
	order.PutUint32(fileData[80:84], summaryRecord)
	order.PutUint32(fileData[84:88], uint32(dataWord))
	copy(fileData[88:96], []byte(format))
	summaryOffset := (summaryRecord - 1) * dafRecordBytes
	putF64(order, fileData[summaryOffset:summaryOffset+8], 0)
	putF64(order, fileData[summaryOffset+8:summaryOffset+16], 0)
	putF64(order, fileData[summaryOffset+16:summaryOffset+24], float64(len(indexed)))
	_ = nameRecord
	for index, segment := range indexed {
		offset := summaryOffset + 24 + index*dafSummaryBytes
		putF64(order, fileData[offset:offset+8], segment.start)
		putF64(order, fileData[offset+8:offset+16], segment.end)
		putI32(order, fileData[offset+16:offset+20], int32(segment.target))
		putI32(order, fileData[offset+20:offset+24], int32(segment.center))
		putI32(order, fileData[offset+24:offset+28], int32(segment.frame))
		putI32(order, fileData[offset+28:offset+32], int32(segment.typ))
		putI32(order, fileData[offset+32:offset+36], int32(segment.startWord))
		putI32(order, fileData[offset+36:offset+40], int32(segment.endWord))
		for word, value := range segment.words {
			putF64(order, fileData[(segment.startWord-1+int64(word))*dafWordBytes:(segment.startWord+int64(word))*dafWordBytes], value)
		}
	}
	path := filepath.Join(t.TempDir(), "test.bsp")
	if err := os.WriteFile(path, fileData, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func putF64(order binary.ByteOrder, destination []byte, value float64) {
	order.PutUint64(destination, math.Float64bits(value))
}

func putI32(order binary.ByteOrder, destination []byte, value int32) {
	order.PutUint32(destination, uint32(value))
}

func assertVector(t *testing.T, got, want [3]float64) {
	assertVectorTolerance(t, got, want, 1e-5)
}

func assertVectorTolerance(t *testing.T, got, want [3]float64, tolerance float64) {
	t.Helper()
	for index := range got {
		if math.Abs(got[index]-want[index]) > tolerance {
			t.Fatalf("vector[%d] = %.12f, want %.12f", index, got[index], want[index])
		}
	}
}

// TestDE441AccuracyVsSkyfield 验收测试：读取真实 DE441 文件，对照 skyfield / acceptance test: read real DE441, compare with skyfield.
func TestDE441AccuracyVsSkyfield(t *testing.T) {
	spkPath := de441Path()
	if _, err := os.Stat(spkPath); os.IsNotExist(err) {
		t.Skip("DE441 file not available; set ASTRO_DE441_BSP")
	}
	refPath := filepath.Join("testdata", "de441_ref.csv")

	spk, err := Open(spkPath)
	if err != nil {
		t.Fatal(err)
	}
	defer spk.Close()

	f, err := os.Open(refPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var maxPosErr, maxVelErr float64
	lineNum := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineNum++
		parts := strings.Split(scanner.Text(), ",")
		if len(parts) != 8 {
			continue
		}
		code, _ := strconv.Atoi(parts[0])
		jd, _ := strconv.ParseFloat(parts[1], 64)
		refPos := [3]float64{}
		refVel := [3]float64{}
		for i := 0; i < 3; i++ {
			refPos[i], _ = strconv.ParseFloat(parts[2+i], 64)
			refVel[i], _ = strconv.ParseFloat(parts[5+i], 64)
		}

		pos, vel, err := spk.State(Body(code), Sun, jd)
		if err != nil {
			t.Fatalf("line %d: State(%d, Sun, %.1f) error: %v", lineNum, code, jd, err)
		}

		for i := 0; i < 3; i++ {
			dPos := math.Abs(pos[i] - refPos[i])
			dVel := math.Abs(vel[i] - refVel[i])
			if dPos > maxPosErr {
				maxPosErr = dPos
			}
			if dVel > maxVelErr {
				maxVelErr = dVel
			}
		}
	}

	// 验收标准：位置 < 100 m，速度 < 0.1 mm/s / acceptance: position < 100 m, velocity < 0.1 mm/s.
	if maxPosErr > 0.1 {
		t.Errorf("max position error = %.6f km (> 100 m)", maxPosErr)
	}
	if maxVelErr > 0.01 { // 0.1 mm/s = 0.01 km/d
		t.Errorf("max velocity error = %.9f km/d (> 0.1 mm/s)", maxVelErr)
	}
	t.Logf("max errors: position %.3e km, velocity %.3e km/d", maxPosErr, maxVelErr)
}

func TestType20IntegralConstantVelocity(t *testing.T) {
	if got, want := chebyshevIntegral([]float64{2}, 0.25), 0.5; got != want {
		t.Fatalf("integral = %.17g, want %.17g", got, want)
	}
}

func TestType21SyntheticRecord(t *testing.T) {
	maxdim := 15
	record := make([]float64, 4*maxdim+12)
	record[0] = float64(maxdim)
	record[1] = 0
	for index := 0; index < maxdim; index++ {
		record[2+index] = 1
	}
	record[maxdim+2] = 10
	record[maxdim+4] = 20
	record[maxdim+6] = 30
	record[maxdim+3] = 1
	record[maxdim+5] = 2
	record[maxdim+7] = 3
	record[4*maxdim+8] = 3
	record[4*maxdim+9] = 1
	record[4*maxdim+10] = 1
	record[4*maxdim+11] = 1
	for component := 0; component < 3; component++ {
		record[8+(component+1)*maxdim] = 1
	}
	gotPos, gotVel, err := evaluateType21(2, record)
	if err != nil {
		t.Fatal(err)
	}
	if gotPos != [3]float64{14, 26, 38} || gotVel != [3]float64{3, 4, 5} {
		t.Fatalf("state = %v, %v", gotPos, gotVel)
	}
}

func TestHorizonsType21CeresAgainstCSPICE(t *testing.T) {
	path := filepath.Join("testdata", "ceres.bsp")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("Horizons Ceres SPK is not available")
	}
	file, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	tests := []struct {
		jd  float64
		pos [3]float64
		vel [3]float64
	}{
		{2461041.5, [3]float64{381172803.95163149, 202285004.74382541, 17811912.880624987}, [3]float64{-8.4262244193126747 * spkSecondsDay, 12.764612139351195 * spkSecondsDay, 7.7363904450095804 * spkSecondsDay}},
		{2461042.5, [3]float64{380442435.16916716, 203386620.60955107, 18480226.044434257}, [3]float64{-8.480449226845364 * spkSecondsDay, 12.735725686923541 * spkSecondsDay, 7.7338055274662123 * spkSecondsDay}},
		{2461043.5, [3]float64{379707384.25882173, 204485733.05701393, 19148311.683061387}, [3]float64{-8.5346068958188273 * spkSecondsDay, 12.706662654284424 * spkSecondsDay, 7.7311236408537631 * spkSecondsDay}},
	}
	for _, test := range tests {
		pos, vel, stateErr := file.State(Body(20000001), Sun, test.jd)
		if stateErr != nil {
			t.Fatal(stateErr)
		}
		assertVectorTolerance(t, pos, test.pos, 1e-6)
		assertVectorTolerance(t, vel, test.vel, 1e-6)
	}
}
