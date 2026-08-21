import { describe, expect, it } from "vitest";

import { formatSize } from "./format-size";

describe("formatSize", () => {
  it.each([
    [512, "512 B"],
    [1024, "1.0 KB"],
    [1024 * 1024, "1.0 MB"],
    [1024 * 1024 * 1024, "1.0 GB"],
  ])("formats %s bytes as %s", (bytes, expected) => {
    expect(formatSize(bytes)).toBe(expected);
  });
});
