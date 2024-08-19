package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/valyala/fastjson"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(300*time.Millisecond))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	checkErr(err)
	resp, err := http.DefaultClient.Do(req)
	checkErr(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.ErrorContext(ctx, fmt.Sprintf("Erro na obtenção da cotação do dólar (status: %d)", resp.StatusCode))
		return
	}

	body, err := io.ReadAll(resp.Body)
	checkErr(err)

	var p fastjson.Parser
	j, err := p.ParseBytes(body)
	checkErr(err)

	os.WriteFile("client\\cotacao.txt", []byte(fmt.Sprintf("Dólar: %s", j.Get("USDBRL").GetStringBytes("bid"))), 0644)
	slog.Info("Cotação salva com sucesso")
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
