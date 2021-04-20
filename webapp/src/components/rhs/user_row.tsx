import React from 'react';

import {useSelector} from 'react-redux';
import {getUser} from 'mattermost-redux/selectors/entities/users';
import {GlobalState} from 'mattermost-redux/types/store';
import {UserProfile} from 'mattermost-redux/types/users';

import {Ownership} from '../../types/badges';

type Props = {
    ownership: Ownership,
    onClick: (user: string) => void
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
        <div>
            <a onClick={() => onClick(ownership.user)}>
                <div>{`@${user.username}`}</div>
                <div>{`Granted by: ${grantedByName}`}</div>
                <div>{`Granted at: ${time.toDateString()}`}</div>
            </a>

        </div>
    );
};

export default UserBadgeRow;
