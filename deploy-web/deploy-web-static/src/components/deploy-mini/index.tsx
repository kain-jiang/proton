import "whatwg-fetch";
import "custom-event-polyfill";
import "core-js/stable/promise";
import "core-js/stable/symbol";
import "core-js/stable/string/starts-with";
import "core-js/web/url";
import React, { FC } from "react";
import { getDefaultAppConfigForDeployMini } from "../../core/bootstrap";
import { FrameWork, Locale } from "@kweaver-ai/workshop-framework-system";

export const DeployMINI: FC<{ lang: Locale }> = React.memo(({ lang }) => {
    const appConfig = getDefaultAppConfigForDeployMini(lang);
    return <FrameWork config={appConfig} />;
});
