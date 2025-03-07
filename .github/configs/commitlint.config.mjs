// Extending the conventional commit rules
export default {
  extends: ["@commitlint/config-conventional"],
  rules: {
    // Enforce commit message length
    "header-max-length": [
      2, // Error level
      128 // max length
    ]
  },
};
