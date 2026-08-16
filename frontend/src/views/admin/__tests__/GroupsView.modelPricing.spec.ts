import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

const source = readFileSync(new URL("../GroupsView.vue", import.meta.url), "utf8");

describe("group model pricing form", () => {
  it("round-trips the long-context switch and model-pricing payload", () => {
    expect(source).toContain("long_context_pricing_enabled");
    expect(source).toContain("model_pricing_json");
    expect(source).toContain("model_pricing: modelPricing");
  });
});
