// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

export const writeTextToClipboard = (text: string, onCopy?: (text: string, result: boolean) => void) => {
    navigator.clipboard.writeText(text)
        .then(() => {
            if (onCopy) onCopy(text, true);
        })
        .catch(() => {
            if (onCopy) onCopy(text, false);
        });
};
