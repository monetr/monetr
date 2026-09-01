package similar

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/elliotcourant/gofx"
	"github.com/monetr/monetr/server/datasources/ofx"
	"github.com/monetr/monetr/server/internal/calc"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/require"
)

// The committed fixtures only produce a vocabulary of a few dozen words, which
// is far too small to show what the vectors look like on an account that has
// been running for years. Point MONETR_OFX_DATASET at a real OFX statement to
// measure the clustering against something representative:
//
//	MONETR_OFX_DATASET=~/Downloads/transactions.ofx \
//	  go test ./server/similar/ -run TestLocalDataset -v
//
// The statement is read from wherever it already lives. Nothing is copied into
// the repository and no transaction text is printed, only aggregates.
const localDatasetEnv = "MONETR_OFX_DATASET"

func loadLocalOFX(t *testing.T, path string) []models.Transaction {
	file, err := os.Open(path)
	require.NoError(t, err, "must be able to open the OFX dataset")
	defer file.Close()

	parsed, err := ofx.ParseFile(file)
	require.NoError(t, err, "must be able to parse the OFX dataset")

	statements := make([]*gofx.StatementTransaction, 0)
	if bank := parsed.BANKMSGSRSV1; bank != nil {
		for _, response := range bank.STMTTRNRS {
			if stmt := response.STMTRS; stmt != nil && stmt.BANKTRANLIST != nil {
				statements = append(statements, stmt.BANKTRANLIST.STMTTRN...)
			}
		}
	}
	if card := parsed.CREDITCARDMSGSRSV1; card != nil {
		for _, response := range card.CCSTMTTRNRS {
			if stmt := response.CCSTMTRS; stmt != nil && stmt.BANKTRANLIST != nil {
				statements = append(statements, stmt.BANKTRANLIST.STMTTRN...)
			}
		}
	}

	// Mirror the mapping the OFX upload job uses so the tokenizer sees exactly
	// what it would see in production.
	out := make([]models.Transaction, 0, len(statements))
	for _, statement := range statements {
		name := strings.TrimSpace(statement.NAME)
		original := strings.TrimSpace(statement.MEMO)
		name = myownsanity.CoalesceStrings(name, original)
		original = myownsanity.CoalesceStrings(original, name)
		if name == "" {
			continue
		}
		date, _ := ofx.ParseDate(statement.DTPOSTED, time.UTC)
		out = append(out, models.Transaction{
			TransactionId:        models.NewID[models.Transaction](),
			Name:                 name,
			OriginalName:         original,
			OriginalMerchantName: name,
			Date:                 date,
		})
	}
	return out
}

func TestLocalDataset(t *testing.T) {
	path := os.Getenv(localDatasetEnv)
	if path == "" {
		t.Skipf("set %s to an OFX statement to run this", localDatasetEnv)
	}

	data := loadLocalOFX(t, path)
	require.NotEmpty(t, data, "the dataset must contain transactions")

	processor := NewTransactionTFIDF()
	for i := range data {
		processor.AddTransaction(&data[i])
	}
	documents := processor.GetDocuments(context.Background())
	require.NotEmpty(t, documents, "the dataset must produce documents")

	vectorSize := len(documents[0].Vector)
	counts := make([]int, 0, len(documents))
	unique := map[string]struct{}{}
	total := 0
	for _, document := range documents {
		counts = append(counts, len(document.Indices))
		total += len(document.Indices)
		unique[fmt.Sprint(document.Vector)] = struct{}{}
	}
	sort.Ints(counts)
	mean := float64(total) / float64(len(counts))

	var pairs, rejected int
	for i := range documents {
		for j := i + 1; j < len(documents); j++ {
			pairs++
			if documents[i].Signature&documents[j].Signature == 0 {
				rejected++
			}
		}
	}

	t.Logf("transactions        %d", len(data))
	t.Logf("documents           %d", len(documents))
	t.Logf("vocabulary          %d", vectorSize)
	t.Logf("nnz                 min=%d p50=%d mean=%.2f p90=%d max=%d",
		counts[0], counts[len(counts)/2], mean, counts[len(counts)*9/10], counts[len(counts)-1])
	t.Logf("density             %.4f%%", 100*mean/float64(vectorSize))
	t.Logf("dense dataset       %.2f MB", float64(len(documents)*vectorSize*4)/1e6)
	t.Logf("sparse dataset      %.3f MB (%.0fx smaller)",
		float64(total*8)/1e6, float64(len(documents)*vectorSize*4)/float64(total*8))
	t.Logf("duplicate vectors   %d unique of %d (%.1fx)",
		len(unique), len(documents), float64(len(documents))/float64(len(unique)))
	t.Logf("signature prefilter rejects %.1f%% of %d pairs",
		100*float64(rejected)/float64(pairs), pairs)

	// The sparse kernel replaced a dense one, so measure both over every pair.
	scratch := make([]float32, vectorSize)
	var sink float32
	start := time.Now()
	for i := range documents {
		for j := range documents {
			sink += calc.EuclideanDistance32(documents[i].Vector, documents[j].Vector)
		}
	}
	dense := time.Since(start)

	start = time.Now()
	for i := range documents {
		for k, index := range documents[i].Indices {
			scratch[index] = documents[i].Values[k]
		}
		for j := range documents {
			dot := calc.SparseDot32(scratch, documents[j].Indices, documents[j].Values)
			sink += documents[i].Norm2 + documents[j].Norm2 - 2*dot
		}
		for _, index := range documents[i].Indices {
			scratch[index] = 0
		}
	}
	sparse := time.Since(start)
	_ = sink

	t.Logf("all-pairs dense     %s", dense)
	t.Logf("all-pairs sparse    %s (%.1fx faster)", sparse, float64(dense)/float64(sparse))

	start = time.Now()
	clusters := NewDBSCAN(documents, Epsilon, MinNeighbors).Calculate(context.Background())
	clustered := 0
	for _, cluster := range clusters {
		clustered += len(cluster.Items)
	}
	t.Logf("DBSCAN              %s -> %d clusters covering %d of %d documents",
		time.Since(start), len(clusters), clustered, len(documents))
}
