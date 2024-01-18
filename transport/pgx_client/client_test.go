package pgx_client

import (
	"context"
	"github.com/pashagolub/pgxmock/v3"
	"testing"
)

// embed pgxmock.PgxConnIface
type PgxPoolMock struct {
	pgxmock.PgxPoolIface
}

func TestClient(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp))

	if err != nil {
		t.Fatal(err)
	}
	client := client{db: &pgDB{conn: mock}}
	row := pgxmock.NewRows([]string{"result"}).AddRow(1)
	mock.ExpectQuery("SELECT 1").WillReturnRows(row)
	var out int
	err = client.DB().QueryRow(context.Background(), "SELECT 1").Scan(&out)
	if err != nil {
		t.Fatal(err)
	}
	if out != 1 {
		t.Errorf("expected 1, got %d", out)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}
