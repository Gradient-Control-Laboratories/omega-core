// Markup for the Bodies side panel, injected by body-panel.js when the app's
// #body-panel is empty. Kept out of each app's index.html so the panel's
// structure lives with the code that drives it, as the Imports and Export
// panels do.

export const BODY_PANEL_MARKUP = `
<div id="body-panel-header">Bodies<span class="panel-close">&#x2715;</span></div>
<div class="panel-body">
    <div id="palette-strip" title="The viewer's palette colours the bodies the document gives no colour. Click to change it."></div>
    <div id="body-rows"></div>
    <input id="material-filter" type="search" placeholder="Filter surfaces by name or description" autocomplete="off">
    <div class="material-keep">Keep
        <label title="Apply only the surface, leaving the body's color. A patterned material's color is its look, so it still sets it."><input id="material-keep-color" type="checkbox"> color</label>
        <label title="Carry the body's displacement (a thread's relief, say) over to the new surface."><input id="material-keep-displacement" type="checkbox"> displacement</label>
        <label title="Carry the body's lay (a brushed or turned grain's direction) over to the new surface."><input id="material-keep-lay" type="checkbox"> lay</label>
    </div>
    <div id="material-readout"></div>
    <div id="material-stack"></div>
    <div id="material-grid"></div>
</div>
`;
