package basic

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/binary"
	"fmt"
	"io"
)

// star_catalog.dat layout after gzip decompression:
// magic[8] | version[1] | rawDataLen[4] | rawData | stringCount[2] |
// repeated(stringLen[uvarint] + stringBytes) |
// maxHR[2] | repeated(maxHR * 6 * stringIndex[2]) | repeated(maxHR * hip[4]).
const starCatalogMagic = "STRCAT01"

//go:embed star_catalog.dat
var starCatalogCompressed []byte

func initStarCatalogData() []byte {
	reader, err := gzip.NewReader(bytes.NewReader(starCatalogCompressed))
	if err != nil {
		panic(err)
	}
	defer reader.Close()
	payload, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	data, detail, hip, err := decodeStarCatalogPayload(payload)
	if err != nil {
		panic(err)
	}
	hr2detail = detail
	hr2hip = hip
	return data
}

func decodeStarCatalogPayload(payload []byte) ([]byte, map[uint16][]string, map[uint16]uint32, error) {
	reader := bytes.NewReader(payload)

	var magic [len(starCatalogMagic)]byte
	if _, err := io.ReadFull(reader, magic[:]); err != nil {
		return nil, nil, nil, fmt.Errorf("read star catalog magic: %w", err)
	}
	if string(magic[:]) != starCatalogMagic {
		return nil, nil, nil, fmt.Errorf("invalid star catalog magic %q", string(magic[:]))
	}

	version, err := reader.ReadByte()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read star catalog version: %w", err)
	}
	if version != 1 {
		return nil, nil, nil, fmt.Errorf("unsupported star catalog version %d", version)
	}

	var rawLen uint32
	if err := binary.Read(reader, binary.LittleEndian, &rawLen); err != nil {
		return nil, nil, nil, fmt.Errorf("read star catalog data length: %w", err)
	}
	data := make([]byte, rawLen)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, nil, nil, fmt.Errorf("read star catalog data: %w", err)
	}

	var stringCount uint16
	if err := binary.Read(reader, binary.LittleEndian, &stringCount); err != nil {
		return nil, nil, nil, fmt.Errorf("read star catalog string count: %w", err)
	}
	stringsTable := make([]string, int(stringCount)+1)
	for i := 1; i < len(stringsTable); i++ {
		length, err := binary.ReadUvarint(reader)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("read star catalog string length: %w", err)
		}
		value := make([]byte, length)
		if _, err := io.ReadFull(reader, value); err != nil {
			return nil, nil, nil, fmt.Errorf("read star catalog string: %w", err)
		}
		stringsTable[i] = string(value)
	}

	var maxHR uint16
	if err := binary.Read(reader, binary.LittleEndian, &maxHR); err != nil {
		return nil, nil, nil, fmt.Errorf("read star catalog max HR: %w", err)
	}
	detail := make(map[uint16][]string)
	for hr := uint16(1); hr <= maxHR; hr++ {
		record := make([]string, 6)
		hasValue := false
		for i := range record {
			var index uint16
			if err := binary.Read(reader, binary.LittleEndian, &index); err != nil {
				return nil, nil, nil, fmt.Errorf("read star catalog detail index: %w", err)
			}
			if int(index) >= len(stringsTable) {
				return nil, nil, nil, fmt.Errorf("detail string index out of range: hr=%d index=%d", hr, index)
			}
			record[i] = stringsTable[index]
			hasValue = hasValue || index != 0
		}
		if hasValue {
			detail[hr] = record
		}
	}

	hip := make(map[uint16]uint32)
	for hr := uint16(1); hr <= maxHR; hr++ {
		var value uint32
		if err := binary.Read(reader, binary.LittleEndian, &value); err != nil {
			return nil, nil, nil, fmt.Errorf("read star catalog HIP: %w", err)
		}
		if value != 0 {
			hip[hr] = value
		}
	}
	return data, detail, hip, nil
}
