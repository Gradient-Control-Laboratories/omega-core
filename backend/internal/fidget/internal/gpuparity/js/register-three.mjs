// Lets Node load the viewers' modules: they import "three", which the pages
// map to the vendored build with an import map. Pass to node with --import.
import { register } from "node:module";

register("./three-resolve.mjs", import.meta.url);
