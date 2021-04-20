// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {Client4} from 'mattermost-redux/client';
import {ClientError} from 'mattermost-redux/client/client4';

import manifest from 'manifest';
import {AllBadgesBadge, BadgeDetails, BadgeID, UserBadge} from 'types/badges';

export default class Client {
    private url: string;

    constructor() {
        this.url = '/plugins/' + manifest.id + '/api/v1';
    }

    async getUserBadges(userID: string): Promise<UserBadge[]> {
        try {
            const res = await this.doGet(`${this.url}/getUserBadges/${userID}`);
            return res as UserBadge[];
        } catch {
            return [];
        }
    }

    async getBadgeDetails(badgeID: BadgeID): Promise<BadgeDetails|null> {
        try {
            const res = await this.doGet(`${this.url}/getBadgeDetails/${badgeID}`);
            return res as BadgeDetails;
        } catch {
            return null;
        }
    }

    async getAllBadges(): Promise<AllBadgesBadge[]> {
        try {
            const res = await this.doGet(`${this.url}/getAllBadges`);
            return res as AllBadgesBadge[];
        } catch {
            return [];
        }
    }

    private doGet = async (url: string, headers: {[x:string]: string} = {}) => {
        headers['X-Timezone-Offset'] = String(new Date().getTimezoneOffset());

        const options = {
            method: 'get',
            headers,
        };

        const response = await fetch(url, Client4.getOptions(options));

        if (response.ok) {
            return response.json();
        }

        const text = await response.text();

        throw new ClientError(Client4.url, {
            message: text || '',
            status_code: response.status,
            url,
        });
    }

    private doPost = async (url: string, body: any, headers: {[x:string]: string} = {}) => {
        headers['X-Timezone-Offset'] = String(new Date().getTimezoneOffset());

        const options = {
            method: 'post',
            body: JSON.stringify(body),
            headers,
        };

        const response = await fetch(url, Client4.getOptions(options));

        if (response.ok) {
            return response.json();
        }

        const text = await response.text();

        throw new ClientError(Client4.url, {
            message: text || '',
            status_code: response.status,
            url,
        });
    }

    private doDelete = async (url: string, headers: {[x:string]: string} = {}) => {
        headers['X-Timezone-Offset'] = String(new Date().getTimezoneOffset());

        const options = {
            method: 'delete',
            headers,
        };

        const response = await fetch(url, Client4.getOptions(options));

        if (response.ok) {
            return response.json();
        }

        const text = await response.text();

        throw new ClientError(Client4.url, {
            message: text || '',
            status_code: response.status,
            url,
        });
    }

    private doPut = async (url: string, body: any, headers: {[x:string]: string} = {}) => {
        headers['X-Timezone-Offset'] = String(new Date().getTimezoneOffset());

        const options = {
            method: 'put',
            body: JSON.stringify(body),
            headers,
        };

        const response = await fetch(url, Client4.getOptions(options));

        if (response.ok) {
            return response.json();
        }

        const text = await response.text();

        throw new ClientError(Client4.url, {
            message: text || '',
            status_code: response.status,
            url,
        });
    }
}
