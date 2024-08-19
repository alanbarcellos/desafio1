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

	f, err := os.OpenFile("client\\cotacao.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	checkErr(err)
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%s - A cotação do dólar é de R$ %s\n", time.Now().Format("02/01/2006 15:04:05"), j.Get("USDBRL").GetStringBytes("bid")))
	checkErr(err)
	slog.Info("Cotação salva com sucesso")
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
