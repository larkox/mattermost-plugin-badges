import React from 'react';

import {UserBadge} from '../../types/badges';
import BadgeImage from '../utils/badge_image';
import {markdown} from 'utils/markdown';

type Props = {
    badge: UserBadge;
    onClick: (badge: UserBadge) => void;
}

const UserBadgeRow: React.FC<Props> = ({badge, onClick}: Props) => {
    const time = new Date(badge.time);
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
            <div>{markdown(badge.description)}</div>
            <div>{`Granted by: ${badge.granted_by_name}`}</div>
            <div>{`Granted at: ${time.toDateString()}`}</div>
        </div>
    );
};

export default UserBadgeRow;
