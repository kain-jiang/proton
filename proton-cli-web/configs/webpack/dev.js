const { merge } = require("webpack-merge");
const commonConfig = require("./common");
const ReactRefreshPlugin = require("@pmmmwh/react-refresh-webpack-plugin");

module.exports = merge(commonConfig, {
  mode: "development",
  devServer: {
    hot: true,
    historyApiFallback: true,
  },
  devtool: "source-map",
  plugins: [new ReactRefreshPlugin()],
});
