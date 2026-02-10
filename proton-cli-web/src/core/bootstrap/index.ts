import favico from "../../assets/img/favicon.ico";

export const setFavicon = () => {
  const head = document.querySelector("head");
  const link = document.createElement("link");
  link.rel = "shortcut icon";
  link.type = "image/x-icon";
  link.href = favico;
  head.appendChild(link);
};

export const setTitle = () => {
  const head = document.querySelector("head");
  const title = document.createElement("title");
  title.innerText = "Proton 部署工具";
  head.appendChild(title);
};
