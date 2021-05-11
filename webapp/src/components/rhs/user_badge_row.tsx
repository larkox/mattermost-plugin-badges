import React from 'react';

import {UserBadge} from '../../types/badges';
import BadgeImage from '../utils/badge_image';
import {markdown} from 'utils/markdown';

import './user_badge_row.scss';

type Props = {
    badge: UserBadge;
    onClick: (badge: UserBadge) => void;
}

const UserBadgeRow: React.FC<Props> = ({badge, onClick}: Props) => {
    const time = new Date(badge.time);
    return (
        <div className='UserBadgesRow'>
            <a onClick={() => onClick(badge)}>
                <span className='user-badge-icon'>
                    <BadgeImage
                        badge={badge}
                        size={32}
                    />
                </span>
            </a>
            <div className='user-badge-text'>
                <div className='user-badge-name'>{badge.name}</div>
                <div className='user-badge-description'>{markdown(badge.description)}</div>
                <div className='badge-type'>{'Type: ' + badge.type_name}</div>
                <div className='user-badge-granted-by'>{`Granted by: ${badge.granted_by_name}`}</div>
                <div className='user-badge-granted-at'>{`Granted at: ${time.toDateString()}`}</div>
            </div>
        </div>
    );
};

export default UserBadgeRow;
