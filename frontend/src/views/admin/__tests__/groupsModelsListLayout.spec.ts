import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, resolve } from "node:path";

import { describe, expect, it } from "vitest";

const currentDir = dirname(fileURLToPath(import.meta.url));
const groupsViewSource = readFileSync(
  resolve(currentDir, "../GroupsView.vue"),
  "utf8",
);
const pricingEntrySource = readFileSync(
  resolve(currentDir, "../../../components/admin/channel/PricingEntryCard.vue"),
  "utf8",
);
const intervalRowSource = readFileSync(
  resolve(currentDir, "../../../components/admin/channel/IntervalRow.vue"),
  "utf8",
);

describe("groups models list layout", () => {
  it("keeps the toolbar outside of the scrolling list content", () => {
    expect(groupsViewSource).toContain("overflow-hidden rounded-lg border");
    expect(groupsViewSource).toContain("max-h-64 space-y-2 overflow-y-auto p-2");
    expect(groupsViewSource).not.toContain("sticky top-0");
  });

  it("keeps group pricing controls responsive in wide dialogs", () => {
    expect(groupsViewSource.match(/width="wide"/g)).toHaveLength(2);
    expect(
      groupsViewSource.match(/flex flex-wrap items-start justify-between gap-3/g),
    ).toHaveLength(2);
    expect(
      groupsViewSource.match(/shrink-0 whitespace-nowrap text-sm text-primary-600/g),
    ).toHaveLength(2);
    expect(pricingEntrySource).toContain("pricing-default-grid");
    expect(pricingEntrySource).toContain(
      "repeat(auto-fit, minmax(8rem, 1fr))",
    );
    expect(intervalRowSource).toContain("pricing-interval-grid");
    expect(intervalRowSource).toContain(
      "repeat(auto-fit, minmax(7.5rem, 1fr))",
    );
  });
});
