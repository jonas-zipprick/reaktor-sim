package board

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
)

const fingerprintPrefix = "b2_"

// tileBytes is the binary size of one encoded tile payload.
const tileBytes = 5

// tileRecordBytes is one sparse record: 1 index byte + tile payload.
const tileRecordBytes = 1 + tileBytes

// Fingerprint encodes the full tile placement state (types, orientations, charges, etc.).
// Only non-empty tiles are stored (as index + payload), so mostly-empty boards
// stay compact. Demands and damage are not included; they come from CLI flags.
func Fingerprint(s *State) string {
	data := encodeTiles(s)
	return fingerprintPrefix + base64.RawURLEncoding.EncodeToString(data)
}

// FromFingerprint reconstructs tile placement from a Fingerprint string.
func FromFingerprint(fp string) (*State, error) {
	if !strings.HasPrefix(fp, fingerprintPrefix) {
		return nil, fmt.Errorf("unbekannter board-fingerprint (erwartet Praefix %q)", fingerprintPrefix)
	}
	data, err := base64.RawURLEncoding.DecodeString(fp[len(fingerprintPrefix):])
	if err != nil {
		return nil, fmt.Errorf("board-fingerprint dekodieren: %w", err)
	}
	if len(data)%tileRecordBytes != 0 {
		return nil, fmt.Errorf("board-fingerprint: %d Bytes, kein Vielfaches von %d", len(data), tileRecordBytes)
	}
	s := NewEmpty()
	n := len(hex.AllBoardCoords)
	for off := 0; off < len(data); off += tileRecordBytes {
		idx := int(data[off])
		if idx >= n {
			return nil, fmt.Errorf("board-fingerprint: Feld-Index %d ausserhalb 0-%d", idx, n-1)
		}
		c := hex.AllBoardCoords[idx]
		tile, err := decodeTile(data[off+1 : off+tileRecordBytes])
		if err != nil {
			return nil, fmt.Errorf("feld (%d,%d): %w", c.Q, c.R, err)
		}
		s.Tiles[c.Q][c.R] = tile
	}
	return s, nil
}

func encodeTiles(s *State) []byte {
	data := make([]byte, 0, tileRecordBytes*4)
	for i, c := range hex.AllBoardCoords {
		t := s.Tiles[c.Q][c.R]
		if t == (field.Tile{}) {
			continue
		}
		data = append(data, byte(i))
		data = append(data, encodeTile(t)...)
	}
	return data
}

func encodeTile(t field.Tile) []byte {
	flags := byte(t.Orientation&0x7)<<1 | boolByte(t.BurnedOut)
	flags |= byte(t.SuperTarget&0x7) << 4
	pending := byte(t.TokamakCounter)
	if t.Type == field.CoalChamber {
		pending = byte(t.UnboundHeat)
	}
	return []byte{
		byte(t.Type),
		flags,
		byte(t.Charge),
		byte(t.StoredVoltage),
		pending,
	}
}

func decodeTile(b []byte) (field.Tile, error) {
	if len(b) != tileBytes {
		return field.Tile{}, fmt.Errorf("ungueltige kachel-Daten")
	}
	typ := field.Type(b[0])
	if typ < field.Empty || typ > field.DistributionStation {
		return field.Tile{}, fmt.Errorf("ungueltiger feldtyp %d", typ)
	}
	tile := field.Tile{
		Type:          typ,
		BurnedOut:     b[1]&1 != 0,
		Orientation:   hex.Rotation((b[1] >> 1) & 0x7),
		SuperTarget:   hex.Rotation((b[1] >> 4) & 0x7),
		Charge:        int(b[2]),
		StoredVoltage: int(b[3]),
	}
	if typ == field.CoalChamber {
		tile.UnboundHeat = int(b[4])
	} else {
		tile.TokamakCounter = int(b[4])
	}
	return tile, nil
}

func boolByte(v bool) byte {
	if v {
		return 1
	}
	return 0
}
