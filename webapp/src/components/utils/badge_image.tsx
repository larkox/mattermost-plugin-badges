import React from 'react';

import {Badge} from '../../types/badges';
import RenderEmoji from '../utils/emoji';
import {IMAGE_TYPE_EMOJI} from '../../constants';

type Props = {
    badge: Badge;
    size: number;
}

const BadgeImage: React.FC<Props> = ({badge, size}: Props) => {
    switch (badge.image_type) {
    case IMAGE_TYPE_EMOJI:
        return (
            <RenderEmoji
                emojiName={badge.image}
                size={size}
            />
        );
    default:
        return null;
    }
};

export default BadgeImage;
