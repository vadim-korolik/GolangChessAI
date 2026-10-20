package analysis

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/location"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/piece"
)

func TestCaptureGenerationMatchesFilteredLegalMoveOrder(t *testing.T) {
	fens := []string{
		"7k/4P3/8/8/8/8/8/K7 w - - 0 1",
		"k6r/6P1/8/8/8/8/8/K7 w - - 0 1",
		"k7/8/8/8/8/8/4p3/7K b - - 0 1",
		"k7/8/8/8/8/8/6p1/K6R b - - 0 1",
		"8/8/8/r4pPK/8/8/8/k7 w - f6 0 1",
		"4k3/5p2/4n3/8/3p4/2B5/8/4R1K1 w - - 0 1",
		"4k3/8/8/8/5n2/8/6r1/6K1 w - - 0 1",
		"r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
		"4k3/8/8/8/8/8/8/R3K2R w KQ - 0 1",
	}
	for _, path := range []string{
		"../../../testdata/abdada_fens.txt",
		"../../../testdata/lichess_2026_10_14_loss_fens.txt",
		"../../../testdata/en_passant_terminal_fens.txt",
		"../../../testdata/checking_passer_fens.txt",
	} {
		file, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(strings.Split(scanner.Text(), "|")[0])
			if line != "" && !strings.HasPrefix(line, "#") {
				fens = append(fens, line)
			}
		}
		if err := scanner.Err(); err != nil {
			t.Fatal(err)
		}
		file.Close()
	}
	for index, fen := range fens {
		for _, cached := range []bool{false, true} {
			t.Run(fmt.Sprintf("position-%d/cache-%t", index, cached), func(t *testing.T) {
				parsed, err := ParseFEN(fen)
				if err != nil {
					t.Fatal(err)
				}
				b := parsed.Board
				b.CacheGetAllMoves = cached
				full := append([]location.Move(nil), (*b.GetAllMovesUnShuffled(parsed.Active, parsed.Previous))...)
				var want []location.Move
				for _, move := range full {
					promotion, _ := move.End.GetPawnPromotion()
					moving := b.GetPiece(move.Start)
					enPassant := moving.GetPieceType() == piece.PawnType && move.Start.GetCol() != move.End.GetCol() && b.IsEmpty(move.End)
					if promotion || !b.IsEmpty(move.End) || enPassant {
						want = append(want, move)
					}
				}
				before := BoardToFEN(b, parsed.Active, parsed.Previous, parsed.FullMove)
				kings := b.KingLocations
				got := b.GetCapturesAndPromotions(parsed.Active, parsed.Previous)
				if !slices.Equal(*got, want) {
					t.Fatalf("capture/promotion list or order differs for %s: got %v, want %v", fen, *got, want)
				}
				if after := BoardToFEN(b, parsed.Active, parsed.Previous, parsed.FullMove); after != before || b.KingLocations != kings {
					t.Fatal("selective generation must restore board state")
				}
				if !slices.Equal(*b.GetAllMovesUnShuffled(parsed.Active, parsed.Previous), full) {
					t.Fatal("selective generation must preserve cached full moves")
				}
				b.CacheGetAllMoves = false
				if !slices.Equal(*b.GetAllMovesUnShuffled(parsed.Active, parsed.Previous), full) {
					t.Fatal("selective generation must restore ordinary generation mode")
				}
			})
		}
	}
}
