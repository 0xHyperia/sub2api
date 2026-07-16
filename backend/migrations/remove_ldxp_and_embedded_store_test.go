package migrations

import (
	"strings"
	"testing"
)

func TestRemoveLdxpAndEmbeddedStoreMigration(t *testing.T) {
	raw, err := FS.ReadFile("180_remove_ldxp_and_embedded_store.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(raw))
	for _, expected := range []string{
		"delete from payment_audit_logs",
		"delete from payment_orders",
		"delete from payment_provider_instances",
		"provider_key = 'ldxp'",
		"payment_card_enabled",
		"purchase_subscription_enabled",
		"purchase_subscription_url",
		"migrated_purchase_subscription",
	} {
		if !strings.Contains(sql, expected) {
			t.Fatalf("migration missing %q", expected)
		}
	}
	if strings.Contains(sql, "drop table") {
		t.Fatal("migration must preserve shared payment tables")
	}
}
