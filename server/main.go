package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	slog.Info("Iniciando aplicação...")
	prepareDB()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	slog.SetDefault(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /cotacao", getCotacaoHandler)
	slog.Info("Iniciando servidor na porta 8080")
	err := http.ListenAndServe(":8080", recovery(mux))
	checkErr(err)
}

func getCotacaoHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(40*time.Millisecond))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	checkErr(err)
	resp, err := http.DefaultClient.Do(req)
	checkErr(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.ErrorContext(r.Context(), fmt.Sprintf("Erro ao obter cotação (status: %d)", resp.StatusCode))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(resp.Body)
	checkErr(err)

	insertCotacao(ctx, string(body))
	w.Write(body)
}

func prepareDB() {
	db := getDB()
	defer db.Close()

	_, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS cotacoes (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
	  date datetime DEFAULT CURRENT_TIMESTAMP,
      data TEXT)`)
	checkErr(err)
}

func insertCotacao(ctx context.Context, data string) {
	ctxExec, cancel := context.WithTimeout(ctx, time.Duration(1*time.Millisecond))
	defer cancel()

	db := getDB()
	defer db.Close()

	_, err := db.ExecContext(ctxExec, `INSERT INTO cotacoes (data) VALUES (?)`, data)
	checkErr(err)
}

func getDB() *sql.DB {
	db, err := sql.Open("sqlite3", "server\\cotacoes.db")
	checkErr(err)
	return db
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rcv := recover(); rcv != nil {
				slog.ErrorContext(r.Context(), "Erro critico:", rcv)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
