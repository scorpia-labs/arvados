// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

import React from 'react';
import { mount } from 'cypress/react';
import { Provider } from 'react-redux';
import { createStore, combineReducers } from 'redux';
import { ThemeProvider } from '@mui/material';
import { CustomTheme } from 'common/custom-theme';
import { ResourceKind } from 'models/resource';

// Import CollectionPanel
import { CollectionPanel } from './collection-panel';

describe('<CollectionPanel />', () => {
    let store;

    const createMockStore = (collection) => {
        const initialState = {
            auth: {
                user: {
                    uuid: 'user-uuid',
                },
            },
            resources: {
                [collection.uuid]: collection,
            },
            router: {
                location: {
                    pathname: `/collections/${collection.uuid}`,
                },
            },
            collectionPanel: {
                item: collection,
            },
            collectionPanelFiles: {},
            detailsCard: {},
            properties: {},
            detailsPanel: {
                resourceUuid: collection.uuid,
            },
        };

        return createStore(combineReducers({
            auth: (state = initialState.auth) => state,
            resources: (state = initialState.resources) => state,
            router: (state = initialState.router) => state,
            collectionPanel: (state = initialState.collectionPanel) => state,
            collectionPanelFiles: (state = initialState.collectionPanelFiles) => state,
            detailsCard: (state = initialState.detailsCard) => state,
            properties: (state = initialState.properties) => state,
            detailsPanel: (state = initialState.detailsPanel) => state,
        }));
    };

    const mockCollection = (isTrashed = false) => ({
        uuid: 'zzzzz-4zz18-0123456789abcde',
        ownerUuid: 'user-uuid',
        createdAt: '2023-01-01T00:00:00.000Z',
        modifiedAt: '2023-01-01T00:00:00.000Z',
        modifiedByUserUuid: 'user-uuid',
        kind: ResourceKind.COLLECTION,
        etag: 'etag',
        name: 'Test Collection',
        description: 'Test Description',
        portableDataHash: '1234567890abcdef1234567890abcdef+100',
        manifestText: '',
        replicationDesired: 2,
        replicationConfirmed: 2,
        replicationConfirmedAt: '2023-01-01T00:00:00.000Z',
        storageClassesDesired: ['default'],
        storageClassesConfirmed: ['default'],
        storageClassesConfirmedAt: '2023-01-01T00:00:00.000Z',
        currentVersionUuid: 'zzzzz-4zz18-0123456789abcde',
        version: 1,
        preserveVersion: false,
        fileCount: 1,
        fileSizeTotal: 100,
        properties: {},
        trashAt: '2023-01-01T00:00:00.000Z',
        deleteAt: '2023-01-01T00:00:00.000Z',
        isTrashed: isTrashed,
    });

    it('renders Overview and Files tabs for a normal collection', () => {
        const collection = mockCollection(false);
        store = createMockStore(collection);
        const match = { params: { id: collection.uuid } };

        mount(
            <Provider store={store}>
                <ThemeProvider theme={CustomTheme}>
                    <CollectionPanel match={match} history={{}} location={{}} />
                </ThemeProvider>
            </Provider>
        );

        cy.get('[data-cy="mpv-tabs"]').within(() => {
            cy.contains('Overview').should('exist');
            cy.contains('Files').should('exist');
        });
    });

    it('renders only Overview tab and omits Files tab for a trashed collection', () => {
        const collection = mockCollection(true);
        store = createMockStore(collection);
        const match = { params: { id: collection.uuid } };

        mount(
            <Provider store={store}>
                <ThemeProvider theme={CustomTheme}>
                    <CollectionPanel match={match} history={{}} location={{}} />
                </ThemeProvider>
            </Provider>
        );

        cy.get('[data-cy="mpv-tabs"]').within(() => {
            cy.contains('Overview').should('exist');
            cy.contains('Files').should('not.exist');
        });
    });
});
