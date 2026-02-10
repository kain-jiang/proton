import * as fs from "fs";
import logger from "../common/logger.js";
import * as multiparty from "multiparty";
import { fetchParse, agent, configData } from "./tools/index.js";

const uploadcert = async (req, res) => {
    const { certType } = req.query;
    let form = new multiparty.Form({
        uploadDir: "/AnyShare/tmp",
    });
    await form.parse(req, async function (err, fields, file) {
        if (err) {
            res.json(err);
            res.status(502);
        } else {
            let result = null;
            const config = configData.Module2Config["deploy-manager"];
            const certKeyPath = file.cert_key[0].path;
            const certCrtPath = file.cert_crt[0].path;
            fs.readFile(certKeyPath, "utf-8", async (err, certKeyFileData) => {
                if (err) {
                    try {
                        fs.unlinkSync(certKeyPath);
                        fs.rmdirSync(file.destination);
                    } catch (err) {}
                    res.json(err);
                    res.end();
                } else {
                    fs.readFile(
                        certCrtPath,
                        "utf-8",
                        async (err, certCrtFileData) => {
                            if (err) {
                                try {
                                    fs.unlinkSync(certCrtPath);
                                    fs.rmdirSync(file.destination);
                                } catch (err) {}
                                res.json(err);
                                res.end();
                            } else {
                                const playload =
                                    config.protocol === "https"
                                        ? {
                                              agent,
                                              method: "PUT",
                                              body: JSON.stringify({
                                                  cert_key: certKeyFileData,
                                                  cert_crt: certCrtFileData,
                                              }),
                                          }
                                        : {
                                              method: "PUT",
                                              body: JSON.stringify({
                                                  cert_key: certKeyFileData,
                                                  cert_crt: certCrtFileData,
                                              }),
                                          };
                                try {
                                    result = await fetchParse(
                                        `${config.protocol}://${config.host}:${config.port}/api/deploy-manager/cert/upload-cert/${certType}`,
                                        playload
                                    );
                                    res.status(200);
                                } catch (err) {
                                    logger.info(
                                        `requst failed: ${req.originalUrl};`
                                    );
                                    logger.info(`error message: ${err}`);
                                    logger.info(
                                        `original service: deploy-service`
                                    );
                                    res.status(500);
                                    result = err;
                                } finally {
                                    res.json(result);
                                }
                            }
                        }
                    );
                }
            });
        }
    });
};

export { uploadcert };
