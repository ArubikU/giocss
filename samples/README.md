# Samples

Each sample lives in its own folder and can be run independently with `go run ./samples/<sample-dir>`.

To regenerate all sample previews plus the root README preview:

```bash
go run ./samples/cmd/generate-previews
```

Every sample stores its screenshot as `preview.png` inside the same folder.

## Gallery

| 01 Hello World | 02 Buttons | 03 Login Form |
| --- | --- | --- |
| ![](sample-01-hello-world/preview.png) | ![](sample-02-buttons/preview.png) | ![](sample-03-login-form/preview.png) |
| 04 Cards | 05 Navigation | 06 Typography |
| ![](sample-04-cards/preview.png) | ![](sample-05-navigation/preview.png) | ![](sample-06-typography/preview.png) |
| 07 Color Swatches | 08 Todo List | 09 Dashboard |
| ![](sample-07-color-swatches/preview.png) | ![](sample-08-todo-list/preview.png) | ![](sample-09-dashboard/preview.png) |
| 10 Dark Theme | 11 Modal | 12 Data Table |
| ![](sample-10-dark-theme/preview.png) | ![](sample-11-modal/preview.png) | ![](sample-12-data-table/preview.png) |
| 13 Tabs | 14 Accordion | 15 Notification Center |
| ![](sample-13-tabs/preview.png) | ![](sample-14-accordion/preview.png) | ![](sample-15-notification-center/preview.png) |
| 16 Side Drawer | 17 Search Autocomplete | 18 Docs Viewer |
| ![](sample-16-side-drawer/preview.png) | ![](sample-17-search-autocomplete/preview.png) | ![](sample-18-docs-viewer/preview.png) |
| 19 Advanced Selectors | 20 Form Rerender | 21 Transparent Todo Board |
| ![](sample-19-advanced-selectors/preview.png) | ![](sample-20-form-rerender/preview.png) | ![](sample-21-transparent-todo-board/preview.png) |

## Highlights

- `sample-12-data-table`: sorting and filtering in a dense data grid.
- `sample-13-tabs`: tab switching with transform rotation support.
- `sample-14-accordion`: expandable sections with rotated chevrons.
- `sample-15-notification-center`: read and dismiss flows for a live feed.
- `sample-16-side-drawer`: fixed drawer, overlay veil, and z-index layering.
- `sample-17-search-autocomplete`: input-driven suggestions and selection state.
- `sample-18-docs-viewer`: state-driven docs route with forms and coverage matrix.
- `sample-19-advanced-selectors`: descendant, child, sibling, and `:nth-child` selector coverage.
- `sample-20-form-rerender`: state-driven form that rerenders while typing without clearing uncontrolled inputs.
- `sample-21-transparent-todo-board`: transparent-window kanban board with drag card affordances.