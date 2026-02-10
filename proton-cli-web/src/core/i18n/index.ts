import i18nfactory from "@anyshare/i18nfactory";
import { languages } from "../language";

export const locale = i18nfactory({
  translations: [...languages],
});
