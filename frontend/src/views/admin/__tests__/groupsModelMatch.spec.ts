import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, resolve } from "node:path";

import { describe, expect, it } from "vitest";

const currentDir = dirname(fileURLToPath(import.meta.url));
const groupsViewSource = readFileSync(
  resolve(currentDir, "../GroupsView.vue"),
  "utf8",
);

describe("groups model match administration", () => {
  it("renders the administrator-owned rule editor for create and update", () => {
    expect(groupsViewSource).toContain('v-model="createForm.model_match_patterns_text"');
    expect(groupsViewSource).toContain('v-model="editForm.model_match_patterns_text"');
    expect(groupsViewSource).toContain("model_match_patterns: modelMatchPatterns");
  });

  it("normalizes rules and blocks an empty rule set before saving", () => {
    expect(groupsViewSource).toContain(".map((pattern) => pattern.trim().toLowerCase())");
    expect(groupsViewSource).toContain("return [...new Set(patterns)].sort();");
    expect(groupsViewSource).toContain('t("admin.groups.modelMatch.required")');
  });
});
