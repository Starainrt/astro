package basic

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

// star_catalog_names.tsv.gz 由同名 .tsv 压缩生成，测试校验两者一致。
//
//go:embed star_catalog_names.tsv.gz
var starCatalogNamesCompressed []byte

// starCatalogNameHeader 单一星名表的列顺序，生成端与加载端必须一致。
var starCatalogNameHeader = []string{"hr", "chinese", "aliases", "bayer_zh", "common", "cst", "cst_zh", "region"}

type starNameRecord struct {
	chinese string
	aliases []string
	bayerZH string
	common  string
	cst     string
	cstZH   string
	region  string
}

func decodeStarNameTable(compressed []byte, starCount int) (map[uint16]starNameRecord, error) {
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("open star name table: %w", err)
	}
	defer reader.Close()
	text, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read star name table: %w", err)
	}
	lines := strings.Split(string(text), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		return nil, errors.New("empty star name table")
	}
	header := strings.Split(strings.TrimSuffix(lines[0], "\r"), "\t")
	if len(header) != len(starCatalogNameHeader) {
		return nil, fmt.Errorf("star name header has %d columns, want %d", len(header), len(starCatalogNameHeader))
	}
	for i, want := range starCatalogNameHeader {
		if header[i] != want {
			return nil, fmt.Errorf("star name header column %d = %q, want %q", i+1, header[i], want)
		}
	}
	names := make(map[uint16]starNameRecord)
	for index, line := range lines[1:] {
		line = strings.TrimSuffix(line, "\r")
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != len(starCatalogNameHeader) {
			return nil, fmt.Errorf("star name line %d has %d columns, want %d", index+2, len(fields), len(starCatalogNameHeader))
		}
		hr, err := strconv.ParseUint(fields[0], 10, 16)
		if err != nil || hr == 0 || int(hr) > starCount {
			return nil, fmt.Errorf("star name line %d has invalid HR %q", index+2, fields[0])
		}
		key := uint16(hr)
		if _, ok := names[key]; ok {
			return nil, fmt.Errorf("star name line %d repeats HR %d", index+2, key)
		}
		chinese := fields[1]
		region := fields[7]
		rec := starNameRecord{
			chinese: chinese,
			aliases: splitStarAliases(fields[2]),
			bayerZH: fields[3],
			common:  fields[4],
			cst:     fields[5],
			cstZH:   fields[6],
			region:  region,
		}
		if chinese == "" && len(rec.aliases) == 0 && rec.bayerZH == "" && rec.common == "" &&
			rec.cst == "" && rec.cstZH == "" && rec.region == "" {
			continue
		}
		names[key] = rec
	}
	if len(names) == 0 {
		return nil, errors.New("empty star name table")
	}
	return names, nil
}

func splitStarAliases(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.FieldsFunc(value, func(r rune) bool { return r == ';' || r == '；' })
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
