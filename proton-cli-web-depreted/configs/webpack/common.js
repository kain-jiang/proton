const { resolve } = require("path");
const HtmlWebpackPlugin = require("html-webpack-plugin");

module.exports = {
  entry: "./index.tsx",
  target: ["browserslist"],
  context: resolve(__dirname, "../../src"),
  resolve: {
    extensions: [".js", ".jsx", ".ts", ".tsx", ".scss", ".sass", ".css"],
  },
  module: {
    rules: [
      {
        test: [/\.jsx?$/, /\.tsx?$/],
        use: ["babel-loader"],
        exclude: /node_modules/,
      },
      {
        test: /\.css$/,
        use: ["style-loader", "css-loader"],
      },
      {
        test: /\.(scss|sass)$/,
        use: ["style-loader", "css-loader", "sass-loader"],
      },
      {
        test: /\.(jpe?g|png|gif|svg|ico)$/,
        loader: "file-loader",
        options: {
          outputPath: "/assets/",
          publicPath: "/assets/",
        },
      },
    ],
  },
  plugins: [
    new HtmlWebpackPlugin({
      template: "index.html",
      inject: true,
    }),
  ],
};
