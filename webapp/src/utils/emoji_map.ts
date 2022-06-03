// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {CustomEmoji} from 'mattermost-redux/types/emojis';

import * as Emoji from './emoji';

export default class EmojiMap {
    private customEmojis: Map<string, CustomEmoji>;
    private customEmojisArray: [string, CustomEmoji][];

    constructor(customEmojis: Map<string, CustomEmoji>) {
        this.customEmojis = customEmojis;

        // Store customEmojis to an array so we can iterate it more easily
        this.customEmojisArray = [...customEmojis];
    }

    has(name: string) {
        return Emoji.EmojiIndicesByAlias.has(name) || this.customEmojis.has(name);
    }

    hasSystemEmoji(name: string) {
        return Emoji.EmojiIndicesByAlias.has(name);
    }

    hasUnicode(codepoint: string) {
        return Emoji.EmojiIndicesByUnicode.has(codepoint);
    }

    get(name: string) {
        if (Emoji.EmojiIndicesByAlias.has(name)) {
            return Emoji.Emojis[Emoji.EmojiIndicesByAlias.get(name)!];
        }

        return this.customEmojis.get(name);
    }

    getUnicode(codepoint: string) {
        return Emoji.Emojis[Emoji.EmojiIndicesByUnicode.get(codepoint)!];
    }

    [Symbol.iterator]() {
        const customEmojisArray = this.customEmojisArray;

        return {
            systemIndex: 0,
            customIndex: 0,
            next() {
                if (this.systemIndex < Emoji.Emojis.length) {
                    const emoji = Emoji.Emojis[this.systemIndex];

                    this.systemIndex += 1;

                    return {value: [(emoji as any).short_names[0], emoji]};
                }

                if (this.customIndex < customEmojisArray.length) {
                    const emoji = customEmojisArray[this.customIndex][1] as any;

                    this.customIndex += 1;
                    const name = emoji.short_name || emoji.name;
                    return {value: [name, emoji]};
                }

                return {done: true};
            },
        };
    }
}
