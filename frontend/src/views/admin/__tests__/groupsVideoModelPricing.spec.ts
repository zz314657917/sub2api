import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

const source = readFileSync(new URL("../GroupsView.vue", import.meta.url), "utf8");

describe("group pricing controls", () => {
  it("keeps model pricing separate from existing image pricing state", () => {
    expect(source).toContain("model_pricing_json");
    expect(source).toContain("syncGroupImageQualityPricing");
  });
});
