// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

import React from 'react';
import { writeTextToClipboard } from './clipboard-action';

export interface CopyToClipboardProps {
    text: string;
    onCopy?: (text: string, result: boolean) => void;
    children: React.ReactNode;
}

export const CopyToClipboard = ({ text, onCopy, children }: CopyToClipboardProps) => {
    const elem = React.Children.only(children) as React.ReactElement<any>;

    const onClick = (event: React.MouseEvent) => {
        writeTextToClipboard(text, onCopy);

        if (typeof elem?.props?.onClick === 'function') {
            elem.props.onClick(event);
        }
    };

    return React.cloneElement(elem, { onClick });
};

export default CopyToClipboard;
