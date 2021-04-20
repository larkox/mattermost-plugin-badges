// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
import React from 'react';
import {useSelector} from 'react-redux';

import {getEmojiImageUrl} from 'mattermost-redux/utils/emoji_utils';
import {GlobalState} from 'mattermost-redux/types/store';

import {getEmojiMap} from 'selectors';

interface ComponentProps {
    emojiName: string;
    size?: number;
    emojiStyle?: React.CSSProperties;
}

const RenderEmoji = ({emojiName, emojiStyle, size}: ComponentProps) => {
    const emojiMap = useSelector((state: GlobalState) => getEmojiMap(state));

    if (!emojiName) {
        return null;
    }

    const emojiFromMap = emojiMap.get(emojiName);
    if (!emojiFromMap) {
        return null;
    }
    const emojiImageUrl = getEmojiImageUrl(emojiFromMap);

    return (
        <span
            className='emoticon'
            data-emoticon={emojiName}
            style={{
                backgroundImage: `url(${emojiImageUrl})`,
                backgroundSize: size,
                height: size,
                width: size,
                maxHeight: size,
                maxWidth: size,
                minHeight: size,
                minWidth: size,
                ...emojiStyle,
            }}
        />
    );
};

RenderEmoji.defaultProps = {
    emoji: '',
    emojiStyle: {},
    size: 16,
};

export default React.memo(RenderEmoji);
