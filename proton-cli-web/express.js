const express = require("express");
const bodyParser = require("body-parser");
const app = express();
const portNumber = 3000;
const sourceDir = "dist";
const { payload } = require("./payload");
let flag = 0;

app.use(express.static(sourceDir));
app
  .use(bodyParser.json({ limit: "5MB" }))
  .use(bodyParser.json({ type: "application/vnd.apache.thrift.json" }))
  .use(bodyParser.urlencoded({ extended: false, limit: "5MB" }))
  .post("/init", (req, res) => {
    res.status(409);
    res.json("/init cannot be called concurrently");
    flag = 0;
  })
  .get("/alpha/result", (req, res) => {
    if (flag < 3) {
      res.status(404);
      res.json("The initialization is running.");
    } else {
      res.status(409);
      res.json("initial cluster fail: %v。");
    }
    flag++;
  })
  .get("/config", (req, res) => {
    res.status(500);
    res.json(payload);
  });
app.get("/success", (req, res) => {
  res.status(200);
  console.log(__dirname);
  res.sendFile("/dist/index.html", { root: __dirname });
});
app.listen(portNumber, () => {
  console.log(`Express web server started: http://localhost:${portNumber}`);
  console.log(`Serving content from /${sourceDir}/`);
});
