package config

import (
	"testing"

	"github.com/joho/godotenv"
)

func TestGetEnvUsesConfiguredValueAndFallback(t *testing.T) {
	const key = "CONCORDIA_TESTE_DB_SETTING"
	t.Setenv(key, "")
	if got := getEnv(key, "TESTE_fallback"); got != "TESTE_fallback" {
		t.Fatalf("getEnv sem valor = %q; esperado fallback", got)
	}
	t.Setenv(key, "TESTE_configurado")
	if got := getEnv(key, "TESTE_fallback"); got != "TESTE_configurado" {
		t.Fatalf("getEnv com valor = %q; esperado valor configurado", got)
	}
}

func TestConnectDatabaseMySQL(t *testing.T) {
	_ = godotenv.Load("../../.env")
	db, err := ConnectDatabase()
	if err != nil {
		t.Skipf("MySQL de integração indisponível; teste ignorado: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("fechamento da conexão MySQL falhou: %v", err)
		}
	})
	if err := db.Ping(); err != nil {
		t.Fatalf("Ping no MySQL falhou: %v", err)
	}
}
