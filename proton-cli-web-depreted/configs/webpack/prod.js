const { resolve } = require("path");
const { merge } = require("webpack-merge");

const commonConfig = require("./common");

module.exports = merge(commonConfig, {
  mode: "production",
  output: {
    filename: "js/bundle.min.js",
    path: resolve(__dirname, "../../dist"),
    publicPath: "/",
  },
});
