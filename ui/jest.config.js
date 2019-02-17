module.exports = {
  setupFilesAfterEnv: ["<rootDir>/setupTests.js"],
  roots: ["<rootDir>/src"],
  transform: {
    "^.+\\.tsx?$": "ts-jest",
  },
  testRegex: "(/__tests__/.*|(\\.|/)(test|spec))\\.tsx?$",
  moduleFileExtensions: ["ts", "tsx", "js", "jsx"],
  verbose: false,
  timers: "fake",
}
