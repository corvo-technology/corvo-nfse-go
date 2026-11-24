package test

import (
	"context"
	"net/http"
	"testing"

	corvonfse "github.com/corvo-technology/corvo-nfse-go"
)

func TestCreditsBalance(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusOK, `{"data": {
			"balance": 200,
			"planRemaining": 150,
			"avulsoBalance": 50,
			"hardLimit": 1000
		}}`)
	}))

	balance, err := client.Credits.Balance(context.Background())
	if err != nil {
		t.Fatalf("credits.balance: %v", err)
	}
	if call.path != "/v1/credits" {
		t.Errorf("expected path /v1/credits, got %s", call.path)
	}
	if balance.Balance != 200 || balance.PlanRemaining != 150 || balance.AvulsoBalance != 50 || balance.HardLimit != 1000 {
		t.Errorf("expected balance decoded, got %+v", balance)
	}
}

func TestCreditsPlansUnwrapsPropsItems(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/credits/plans" {
			t.Errorf("expected plans path, got %s", r.URL.Path)
		}
		writeJSON(w, http.StatusOK, `{
			"data": [{
				"id": "plan_1", "slug": "equipe", "name": "Equipe",
				"dailyLimit": 300, "creditUnitPriceCents": 7,
				"supportTier": "standard", "active": true
			}],
			"pagination": {"limit": 50, "hasMore": false, "nextCursor": null}
		}`)
	}))

	plans, _, err := client.Credits.Plans(context.Background(), nil)
	if err != nil {
		t.Fatalf("credits.plans: %v", err)
	}
	if len(plans) != 1 || plans[0].Slug != "equipe" || plans[0].CreditUnitPriceCents != 7 {
		t.Errorf("expected plan decoded, got %+v", plans)
	}
}

func TestCreditsIterateHistoryWalksPropsPages(t *testing.T) {
	page := 0
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		if page == 1 {
			writeJSON(w, http.StatusOK, `{
				"data": [{"id": "tx_1", "type": "consumption", "amount": 1, "balanceAfter": 199}],
				"pagination": {"limit": 1, "hasMore": true, "nextCursor": "h1"}
			}`)
			return
		}
		writeJSON(w, http.StatusOK, `{
			"data": [{"id": "tx_2", "type": "consumption", "amount": 1, "balanceAfter": 198}],
			"pagination": {"limit": 1, "hasMore": false, "nextCursor": null}
		}`)
	}))

	var seen []string
	err := client.Credits.IterateHistory(context.Background(), &corvonfse.ListParams{Limit: 1},
		func(tx corvonfse.CreditTransaction) bool {
			seen = append(seen, tx.ID)
			return true
		})
	if err != nil {
		t.Fatalf("credits.iterateHistory: %v", err)
	}
	if len(seen) != 2 || seen[0] != "tx_1" || seen[1] != "tx_2" {
		t.Errorf("expected both ledger entries, got %v", seen)
	}
	if page != 2 {
		t.Errorf("expected 2 pages, got %d", page)
	}
}

func TestBillingPlanUsesSnakeCaseWireNames(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusOK, `{"data": {
			"plan": {
				"slug": "solo", "name": "Solo", "monthly_allowance": 300,
				"daily_limit": 300, "max_active_companies": 5, "monthly_price_cents": 4990
			},
			"subscription": {
				"status": "active", "paid_until": "2026-11-01",
				"cancel_at_period_end": false, "scheduled_plan": null, "blocked": false,
				"stripe_portal_available": true, "stripe_subscription": true,
				"pix_subscription": {"brCode": "000201", "paymentLinkUrl": null, "planSlug": "solo", "status": "pending"}
			},
			"usage": {
				"avulso_balance": 7, "hard_limit": null,
				"credits_used": 100, "credits_reserved": 0, "credits_remaining": 200,
				"monthly_allowance": 300, "daily_usage": 12, "daily_limit": 300,
				"active_companies": 3, "max_active_companies": 5,
				"period_start": "2026-10-01", "period_end": "2026-11-01"
			}
		}}`)
	}))

	plan, err := client.Billing.Plan(context.Background())
	if err != nil {
		t.Fatalf("billing.plan: %v", err)
	}
	if call.path != "/v1/billing/plan" {
		t.Errorf("expected path /v1/billing/plan, got %s", call.path)
	}
	if plan.Plan.Slug != "solo" || plan.Plan.MonthlyAllowance != 300 || plan.Plan.MonthlyPriceCents != 4990 {
		t.Errorf("expected snake_case plan fields decoded, got %+v", plan.Plan)
	}
	if plan.Subscription.Status != "active" || plan.Subscription.PaidUntil != "2026-11-01" {
		t.Errorf("expected subscription decoded, got %+v", plan.Subscription)
	}
	if !plan.Subscription.StripePortalAvailable || !plan.Subscription.StripeSubscription ||
		plan.Subscription.PixSubscription == nil || plan.Subscription.PixSubscription.Status != "pending" {
		t.Errorf("expected stripe and pix subscription fields decoded, got %+v", plan.Subscription)
	}
	if plan.Usage.AvulsoBalance != 7 {
		t.Errorf("expected usage.avulso_balance 7, got %d", plan.Usage.AvulsoBalance)
	}
	if plan.Usage.CreditsRemaining != 200 || plan.Usage.HardLimit != 0 {
		t.Errorf("expected usage decoded (hard_limit null -> 0), got %+v", plan.Usage)
	}
}

func TestBillingUsage(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/billing/usage" {
			t.Errorf("expected path /v1/billing/usage, got %s", r.URL.Path)
		}
		writeJSON(w, http.StatusOK, `{"data": {
			"canEmit": true,
			"creditsRemaining": 200,
			"creditsUsed": 100,
			"creditsReserved": 0,
			"monthlyAllowance": 300,
			"dailyUsage": 12,
			"dailyLimit": 300,
			"activeCompanies": 3,
			"maxActiveCompanies": 5,
			"periodStart": "2026-10-01",
			"periodEnd": "2026-11-01"
		}}`)
	}))

	quota, err := client.Billing.Usage(context.Background())
	if err != nil {
		t.Fatalf("billing.usage: %v", err)
	}
	if !quota.CanEmit || quota.CreditsRemaining != 200 || quota.MaxActiveCompanies != 5 {
		t.Errorf("expected quota decoded, got %+v", quota)
	}
}
