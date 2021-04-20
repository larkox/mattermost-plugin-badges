import React from 'react';

import {AllBadgesBadge, UserBadge} from '../../types/badges';
import {IMAGE_TYPE_EMOJI} from '../../constants';
import BadgeImage from '../utils/badge_image';

type Props = {
    badge: AllBadgesBadge
    onClick: (badge: AllBadgesBadge) => void
}

function getGrantedText(badge: AllBadgesBadge): string {
    if (badge.granted === 0) {
        return "Not yet granted."
    }
    if (badge.multiple) {
        return `Granted ${badge.granted_times} to ${badge.granted} users.`
    }

    return `Granted to ${badge.granted} users.`
}

const AllBadgesRow: React.FC<Props> = ({badge, onClick}: Props) => {
    return (
        <div>
            <a onClick={() => onClick(badge)}>
                <span>
                    <BadgeImage
                        badge={badge}
                        size={32}
                    />
                </span>
            </a>
            <div>{badge.name}</div>
            <div>{badge.description}</div>
            <div>{getGrantedText(badge)}</div>
        </div>
    );
};

export default AllBadgesRow;
