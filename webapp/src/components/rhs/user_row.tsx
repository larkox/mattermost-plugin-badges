import React from 'react';

import {useSelector} from 'react-redux';
import {getUser} from 'mattermost-redux/selectors/entities/users';
import {GlobalState} from 'mattermost-redux/types/store';
import {UserProfile} from 'mattermost-redux/types/users';

import {Ownership} from '../../types/badges';

import './user_row.scss';
type Props = {
    ownership: Ownership;
    onClick: (user: string) => void;
}

const UserBadgeRow: React.FC<Props> = ({ownership, onClick}: Props) => {
    const user = useSelector<GlobalState, UserProfile>((state) => getUser(state, ownership.user));
    const grantedBy = useSelector<GlobalState, UserProfile>((state) => getUser(state, ownership.granted_by));

    if (!user) {
        return null;
    }

    let grantedByName = 'unknown';
    if (grantedBy) {
        grantedByName = '@' + grantedBy.username;
    }

    const time = new Date(ownership.time);
    return (
        <div className='UserRow'>
            <div className='badge-user-username'><a onClick={() => onClick(ownership.user)}>{`@${user.username}`}</a></div>
            <div className='badge-user-granted-by'>{`Granted by: ${grantedByName}`}</div>
            <div className='badge-user-granted-at'>{`Granted at: ${time.toDateString()}`}</div>
        </div>
    );
};

export default UserBadgeRow;
