import local from "../local";
import session from "../session";

export const lang = {
  /**
   * 中文
   */
  ZHCN: "zh-cn",
  /**
   * 繁体
   */
  ZHTW: "zh-tw",
  /**
   * 英文
   */
  ENUS: "en-us",
};

export const languages = Object.values(lang);

/**
 * 语言资源列表
 */
export const Languages = [
  {
    language: "zh-cn",
    title: "简体中文",
  },
  {
    language: "zh-tw",
    title: "繁體中文",
  },
  {
    language: "en-us",
    title: "English",
  },
];

/**
 * 设置语言
 * @param language 语言
 */
export const setLanguage = (language: string) => {
  session.set("lang", language.toLowerCase());
};

/**
 * 从Hash参数中获取语言
 * @returns {*}
 */
export function getLanguageHash() {
  let hash = window.location.hash;
  let match = /\blang=([a-zA-Z-]+)\b/.exec(hash);
  let lang;

  if (match) {
    lang = match[1];
    setLanguage(lang);

    return lang;
  }
}

/**
 * 获取当前的语言
 * @returns {Object} 返回当前语言
 */
export function getEnvLanguage(): string {
  return (
    getLanguageHash() ||
    session.get("lang") ||
    local.get("lang") ||
    window.navigator["userLanguage"] ||
    window.navigator["language"] ||
    window.navigator["browserLanguage"] ||
    Languages[0].language
  )
    .trim()
    .toLowerCase();
}

/**
 * 通过语言环境获取翻译
 * @param textZHCN 中文翻译
 * @param textZHTW 繁体翻译
 * @param textENUS 英文翻译
 * @returns 翻译
 */
export const getLocaleByEnv = (
  textZHCN: string,
  textZHTW: string,
  textENUS: string,
  language?: string,
) => {
  const lang = language ? language : getEnvLanguage();
  switch (lang) {
    case Languages[0].language:
      return textZHCN;
    case Languages[1].language:
      return textZHTW;
    default:
      return textENUS;
  }
};
