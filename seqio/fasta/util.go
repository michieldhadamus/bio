package fasta

import (
	"errors"
	"runtime"
	"strings"

	"github.com/shenwei356/bio/seq"
	"github.com/shenwei356/breader"
)

// GetSeqNames returns the names of a fasta file
func GetSeqNames(file string) ([]string, error) {
	names := []string{}

	fn := func(line string) (interface{}, bool, error) {
		if len(line) == 0 {
			return "", false, nil
		}
		if line[0] == '>' {
			line = strings.TrimRight(line, "\n")
			if len(line) > 1 {
				return line[1:], true, nil
			}
			return "", true, nil
		}
		return "", false, nil
	}
	reader, err := breader.NewBufferedReader(file, runtime.NumCPU(), 100, fn)
	if err != nil {
		return names, err
	}

	for chunk := range reader.Ch {
		for _, data := range chunk.Data {
			names = append(names, data.(string))
		}
	}
	return names, nil
}

// GetSeqs return fasta records of a file.
// when alphabet is nil or seq.Unlimit, it will automaticlly detect the alphabet.
// when idRegexp is "", default idRegexp ( ^([^\s]+)\s? ) will be used.
func GetSeqs(file string, alphabet *seq.Alphabet, chunkSize int, threads int, idRegexp string) ([]*FastaRecord, error) {
	records := []*FastaRecord{}
	if alphabet == nil || alphabet == seq.Unlimit {
		alphabet = nil
	}
	fastaReader, err := NewFastaReader(alphabet, file, threads, chunkSize, idRegexp)
	if err != nil {
		return records, err
	}
	for chunk := range fastaReader.Ch {
		if err != nil {
			return records, err
		}

		for _, record := range chunk.Data {
			records = append(records, record)
		}
	}
	return records, nil
}

// GetSeqsMap returns all seqs as a map for fasta file
func GetSeqsMap(file string, alphabet *seq.Alphabet, chunkSize int, threads int, idRegexp string) (map[string]*FastaRecord, error) {
	m := make(map[string]*FastaRecord)
	records, err := GetSeqs(file, alphabet, chunkSize, threads, idRegexp)
	if err != nil {
		return m, err
	}
	for _, record := range records {
		m[string(record.Name)] = record
	}
	return m, nil
}

// GuessAlphabet guess the alphabet of the file
func GuessAlphabet(file string) (*seq.Alphabet, error) {
	fastaReader, err := NewFastaReader(seq.Unlimit, file, 0, 1, "")
	if err != nil {
		return seq.Unlimit, err
	}

	for {
		select {
		case chunk := <-fastaReader.Ch:
			if chunk.Err != nil {
				return seq.Unlimit, chunk.Err
			}
			if len(chunk.Data) == 0 {
				return seq.Unlimit, errors.New("no fasta records found in file: " + file)
			}
			firstRecord := chunk.Data[0]
			fastaReader.Cancel()
			return seq.GuessAlphabet(firstRecord.Seq.Seq), nil
		default:
		}
	}
}
