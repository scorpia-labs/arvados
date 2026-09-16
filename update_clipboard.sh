#!/bin/bash

# Remove unnecessary checks for navigator.clipboard
sed -i 's/if (navigator.clipboard && navigator.clipboard.writeText) {/if (navigator.clipboard) {/g' services/workbench2/src/components/copy-to-clipboard/copy-result-to-clipboard.ts
sed -i 's/if (navigator.clipboard && navigator.clipboard.writeText) {/if (navigator.clipboard) {/g' services/workbench2/src/components/copy-to-clipboard/copy-to-clipboard.tsx
sed -i 's/if (textToCopy && navigator.clipboard && navigator.clipboard.writeText) {/if (textToCopy && navigator.clipboard) {/g' services/workbench2/src/store/open-in-new-tab/open-in-new-tab.actions.ts
sed -i 's/if (string.length && navigator.clipboard && navigator.clipboard.writeText) {/if (string.length && navigator.clipboard) {/g' services/workbench2/src/store/open-in-new-tab/open-in-new-tab.actions.ts
sed -i 's/if (props.href && navigator.clipboard && navigator.clipboard.writeText) {/if (props.href && navigator.clipboard) {/g' services/workbench2/src/views-components/context-menu/actions/copy-to-clipboard-action.tsx

# Simplify optional chaining check in copy-to-clipboard.tsx
sed -i "s/if (elem && elem.props && typeof elem.props.onClick === 'function') {/if (typeof elem?.props?.onClick === 'function') {/g" services/workbench2/src/components/copy-to-clipboard/copy-to-clipboard.tsx
sed -i "s/if (elem && elem.props && typeof elem.props.onClick === 'function') {/if (typeof elem?.props?.onClick === 'function') {/g" services/workbench2/src/components/copy-to-clipboard/copy-result-to-clipboard.ts
