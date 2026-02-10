import {
    login,
    logout,
    refreshToken,
    oauthLoginCallback,
    oauthLogoutCallback,
    getUserInfoByToken,
} from "../handlers/auth";
import { meta } from "../handlers/meta";
import { uploadcert } from "../handlers/uploadcert";
import { restfulProxy, interfaceProxy } from "../handlers/proxyroutes";
import { getOemconfig } from "../handlers/oemconfig";

export const resgisterRouting = (app) => {
    app.get("/interface/deployweb/login", login);
    app.get("/interface/deployweb/oauth/login/callback", oauthLoginCallback);
    app.head("/interface/deployweb/meta", meta);
    // interfaceProxy is for authentication, the routes written above this line are authenticated
    app.all("/interface/deployweb/*", interfaceProxy);
    app.post("/interface/deployweb/logout", logout);
    app.get("/interface/deployweb/refreshtoken", refreshToken);
    app.get("/interface/deployweb/oauth/logout/callback", oauthLogoutCallback);
    app.get(
        "/interface/deployweb/oauth/getUserInfoByToken",
        getUserInfoByToken
    );
    app.get("/api/deploy-web-service/v1/oemconfig", getOemconfig);
    app.all(
        ["/api/deployweb/deploy-manager/*", "/api/deployweb/audit-log/*"],
        restfulProxy
    );
    app.put("/interface/deployweb/upload/cert/", uploadcert);
};
