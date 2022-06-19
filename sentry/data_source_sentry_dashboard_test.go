package sentry

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/jianyuan/go-sentry/v2/sentry"
)

func TestAccSentryDashboardDataSource_basic(t *testing.T) {

	dashboardTitle := acctest.RandomWithPrefix("tf-dashboard")
	rn := "sentry_dashboard.test"
	dn := "data.sentry_dashboard.test"
	rnCopy := "sentry_dashboard.test_copy"

	check := func(name, dashboardTitle string) resource.TestCheckFunc {
		var dashboard sentry.Dashboard

		return resource.ComposeTestCheckFunc(
			testAccCheckSentryDashboardExists(name, &dashboard),
			resource.TestCheckResourceAttrWith(name, "internal_id", func(v string) error {
				want := sentry.StringValue(dashboard.ID)
				if v != want {
					return fmt.Errorf("got dashboard ID %s; want %s", v, want)
				}
				return nil
			}),
			resource.TestCheckResourceAttr(name, "organization", testOrganization),
			resource.TestCheckResourceAttr(name, "title", dashboardTitle),
			resource.TestCheckResourceAttr(name, "widget.#", "1"),
			resource.TestCheckTypeSetElemNestedAttrs(name, "widget.*", map[string]string{
				"title":        "Custom Widget",
				"display_type": "world_map",
			}),
			resource.TestCheckResourceAttr(name, "widget.0.query.#", "1"),
			resource.TestCheckTypeSetElemNestedAttrs(name, "widget.0.query.*", map[string]string{
				"name":       "Metric",
				"conditions": "!event.type:transaction",
			}),
			resource.TestCheckResourceAttr(name, "widget.0.query.0.fields.#", "1"),
			resource.TestCheckResourceAttr(name, "widget.0.query.0.fields.0", "count()"),
			resource.TestCheckResourceAttr(name, "widget.0.query.0.aggregates.#", "1"),
			resource.TestCheckResourceAttr(name, "widget.0.query.0.aggregates.0", "count()"),
			resource.TestCheckResourceAttr(name, "widget.0.layout.#", "1"),
			resource.TestCheckTypeSetElemNestedAttrs(name, "widget.0.layout.*", map[string]string{
				"x":     "0",
				"y":     "0",
				"w":     "2",
				"h":     "1",
				"min_h": "1",
			}),
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSentryDashboardDataSourceConfig(dashboardTitle),
				Check: resource.ComposeTestCheckFunc(
					check(rn, dashboardTitle),
					check(dn, dashboardTitle),
					check(rnCopy, dashboardTitle+"-copy"),
				),
			},
		},
	})
}

func testAccSentryDashboardDataSourceConfig(dashboardTitle string) string {
	return fmt.Sprintf(`
data "sentry_organization" "test" {
	slug = "%[1]s"
}

resource "sentry_dashboard" "test" {
	organization = data.sentry_organization.test.id
	title        = "%[2]s"

	widget {
		title        = "Custom Widget"
		display_type = "world_map"

		query {
			name       = "Metric"

			fields     = ["count()"]
			aggregates = ["count()"]
			conditions = "!event.type:transaction"
		}

		layout {
			x     = 0
			y     = 0
			w     = 2
			h     = 1
			min_h = 1
		}
	}
}

data "sentry_dashboard" "test" {
	organization = sentry_dashboard.test.organization
	internal_id  = sentry_dashboard.test.internal_id
}

resource "sentry_dashboard" "test_copy" {
	organization = data.sentry_dashboard.test.organization
	title        = "${data.sentry_dashboard.test.title}-copy"

	dynamic "widget" {
		for_each = data.sentry_dashboard.test.widget
		content {
			title        = widget.value.title
			display_type = widget.value.display_type
			interval     = widget.value.interval
			widget_type  = widget.value.widget_type
			limit        = widget.value.limit

			dynamic "query" {
				for_each = widget.value.query
				content {
					name = query.value.name

					fields        = query.value.fields
					aggregates    = query.value.aggregates
					columns       = query.value.columns
					field_aliases = query.value.field_aliases
					conditions    = query.value.conditions
					order_by      = query.value.order_by
				}
			}

			layout {
				x     = widget.value.layout[0].x
				y     = widget.value.layout[0].y
				w     = widget.value.layout[0].w
				h     = widget.value.layout[0].h
				min_h = widget.value.layout[0].min_h
			}
		}
	}
}
	`, testOrganization, dashboardTitle)
}
