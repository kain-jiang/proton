import { TextInterface } from './declare'
import Text from './text'
import { Tips } from './tips'
import { Dot } from './dot'

(Text as any).Tips = Tips;
(Text as any).Dot = Dot;

export default Text as TextInterface
