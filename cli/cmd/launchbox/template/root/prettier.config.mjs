/**
 * @type {import("prettier").Config}
 */
const prettierConfig = {
  printWidth: 80,
  semi: false,
  singleAttributePerLine: true,
  singleQuote: false,
  trailingComma: "all",
  importOrder: [
    "^react",
    "^next",
    "<THIRD_PARTY_MODULES>",
    "^@/(?!components/).*",
    "^@/components/.*",
    "^[./]",
  ],
  importOrderSeparation: true,
  importOrderSortSpecifiers: true,
  plugins: ["@trivago/prettier-plugin-sort-imports"],
}

export default prettierConfig
