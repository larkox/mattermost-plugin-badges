export type Badge = {
    id: BadgeID;
    name: string;
    description: string;
    image: string;
    image_type: BadgeImageType;
    multiple: boolean;
    type: BadgeType;
    created_by: string;
}

export type Ownership = {
    user: string;
    granted_by: string;
    badge: BadgeID;
    time: number;
}

export type BadgeID = number;
export type BadgeType = number
export type BadgeImageType = string;

export type UserBadge = Badge & Ownership & {granted_by_name: string};
export type BadgeDetails = Badge & {
    owners: OwnershipList,
    created_by_username: string,
}
export type AllBadgesBadge = Badge & {
    granted: number,
    granted_times: number,
}

export type OwnershipList = Ownership[]

export type BadgeTypeDefinition = {
    id: BadgeType;
    name: string;
    frame: string;
}
