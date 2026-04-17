declare module "*.ico";
declare module "*.png";
declare module "*.jpg";
declare module "*.jpeg";
declare module "*.bmp";
declare module "*.gif";
declare module "*.avif";

declare module "*.svg" {
  export const ReactComponent: React.FunctionComponent<
    React.SVGAttributes<SVGElement>
  >;
}
declare module "*.sass" {
  const content: Record<string, string>;
  export default content;
}

declare module "*.scss" {
  const content: Record<string, string>;
  export default content;
}

declare module "*.less" {
  const content: Record<string, string>;
  export default content;
}

declare module "*.css" {
  const content: Record<string, string>;
  export default content;
}
