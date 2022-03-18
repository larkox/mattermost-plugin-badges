# Badges plugin
Let your users show appreciation for their fellow colleagues by granting badges.

## Install
Get the latest release from [GitHub](https://github.com/larkox/mattermost-plugin-badges/releases) and [install it manually](https://developers.mattermost.com/integrate/plugins/server/hello-world/#installing-the-plugin) on your server.

## Configuration

![Screenshot from 2022-03-16 11-02-13](https://user-images.githubusercontent.com/1933730/158565396-9d637c4c-6772-449f-81cb-2b73f8f6670e.png)

- **Badges admin**: Every System Admin is considered a badges admin. System Admins can assign the badges admin role to a single person by specifying their username. Only a single badge admin assignment is permitted.

## Usage
### Creating a type
Badge admins can create different types of badges, and each type of badge can have its own permissions. You must be a badge admin to create a badge type.
Run the slash command `/badges create type` to open the creation dialog.

![Screenshot from 2022-03-16 11-14-31](https://user-images.githubusercontent.com/1933730/158567578-1241cc93-6964-4dc7-a56b-a5b3729229b7.png)

- **Name**: The type of badge that's visible in the badges description.
- **Everyone can create badge**: If you mark this checkbox, every user in your Mattermost instance can create badges of this type.
- **Can create allowlist**: This list contains the usernames (comma separated) of all the people allowed to create badges of this type.
- **Everyone can grant badge**: If you mark this checkbox, every user in your Mattermost instance can grant any badge of this type.
- **Can grant allowlist**: this list contains the usernames (comma separated) of all the people allowed to grant badges of this type.

### Permissions details
Badge admins can always create types, create badges for any type, and grant badges from any type, regardless of the permissions in place for a given badge type.
A badge creator can always grant the badge they created.
Any other user is subject to the permissions defined as part of the badge type.

Some examples of badge permissions by type are included below. Remember that badge admins have full control over badges, and badge creators can always grant badges. The examples below are intended to demonstrate how badge permissions can be configured for non-admin users to get the most out of badges. 
(ECC: Everyone Can Create, CC: Can Create Allowlist, ECG: Everyone Can Grant, CG: Can Grant Allowlist)
| Permissions                                                                                  | Example                                                                                                                                                 | ECC   | ECG   | CC           | CG           |
|----------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|-------|-------|--------------|--------------|
| Only badge admins can create and grant badges                                                |                                                                                                                                                         | false | false | empty        | empty        |
| Only user1 can create badges, but everyone can grant them                                    | for peer appreciation badges, like "Thank you" badge                                                                                                    | false | true  | user1        | empty        |
| Only user1 can create badges, and only user2 and user3 can grant them                        | lead appreciation badges, like "MVP" badge, where the management create the badges, and the team leads are the ones granting them to their team members | false | false | user1        | user2, user3 |
| Only user1 and user2 can create badges, but they can only grant the badges they have created | can be used to have team specific badges without creating a new type for every team                                                                     | false | false | user1, user2 | empty        |
| Everyone can create badges, but can only grant the badges they have created                  |                                                                                                                                                         | true  | false | empty        | empty        |
| Everyone can create and grant any badge                                                      |                                                                                                                                                         | true  | true  | empty        | empty        |

### Creating a badge
Run the slash command `/badges create badge` to open the creation dialog.

![Screenshot from 2022-03-16 11-37-32](https://user-images.githubusercontent.com/1933730/158571687-4983f7e4-1cf9-4fa1-a3f1-6f80918e28e3.png)

- **Name**: Name of the badge.
- **Description**: Description of the badge.
- **Image**: Only emojis are allowed. You must input the emoji name as you would do to add it to a message (e.g. `:+1:` or `:smile:`). Custom emojis are also allowed.
- **Type**: The type of badge. This list will show only types you have permissions to create.
- **Multiple**: Whether this badge can be granted more than once to the same person.

### Details about Multiple
All badges can be assigned to any number of people. What the **Multiple** setting controls is whether this badge can be granted more than once to the same person. For example, a "Thank you" badge should be grantable many times (many people can be thankful to you on more than one occasion), and therefore, a Thank You badge should have the **Multiple** option selected. However, a "First year in the company" badge should be granted only once since a user won't celebrate this milestone multiple times at the same company. This type of badge should have the **Multiple** option unselected.

### Granting a badge
There are two ways to open the grant dialog:
- Run the `/badges grant` command.
- Click on the "Grant badge" link available in the Profile Popover, visible when you click on someone's username.

![Screenshot from 2022-03-16 11-47-14](https://user-images.githubusercontent.com/1933730/158573673-723e77a2-6d58-4aa5-8a89-6adcbce50e13.png)

The dialog looks like this:

![Screenshot from 2022-03-16 11-51-05](https://user-images.githubusercontent.com/1933730/158573834-70ea72b0-4a03-4b09-a694-751c0ca1ba04.png)

- **User**: The user you want to grant the badge to (may be prepopulated if you clicked the grant button from the profile popover, or added the username in the command).
- **Badge**: The badge you want to grant (may be prepopulated if you added the badge id in the command).
- **Reason**: An optional reason why you are awarding this badge. (Specially useful for badges like "Thank you").
- **Notify on this channel**: If you select this option, a message from the badges bot will be posted in the current channel, letting everyone in that channel know that you granted this badge to that person.
The user that received the badge will always receive a DM from the badges bot letting them know they have been awarded a badge. In addition, the following may happen:
- If **Notify on this channel** was marked, the badges bot will post a message on the current channel letting everyone know that the user has been awarded a badge.
- If a subscription for this badge type is set, the badges bot will post a message on all subscribed channels letting everyone know that the user has been awarded a badge.

![Screenshot from 2022-03-16 12-28-45](https://user-images.githubusercontent.com/1933730/158580318-592bb139-6c43-48f0-99c3-79d868aa8024.png)

If you try to award a badge that can't be awarded more than once to a single recipient, the badge won't be granted..

### Subscriptions
In order to create a subscription, you must be a badges admin.
Subscriptions will create posts into a channel every time a badge is granted. There is no limit to the number of subscriptions per channel or per type.
There are two ways to open the subscription creation dialog:
- Run the `/badges subscription create` command.
- Click on the **Add badge subscription** menu from the channel menu.

![Screenshot from 2022-03-16 12-16-16](https://user-images.githubusercontent.com/1933730/158578166-1ae6f5de-a53b-4e46-95ba-4fd57f50a315.png)

The dialog looks like this:

![Screenshot from 2022-03-16 12-16-55](https://user-images.githubusercontent.com/1933730/158578272-dc6644a1-3a8b-4f54-8c83-d192d8fab273.png)

- **Type**: The type of badges you want to subscribe to this channel.

In order to remove subscriptions, a similar dialog can be opened by using the `/badges subscription remove` and the "Remove badge subscription" menu from the channel menu.

### Editing a deleting badges and types
In order to edit or delete types you must be a badge admin. In order to edit or delete a badge, you must be a badge admin or the creator.
Run `/badges edit type --type typeID` or `/badges edit badge --id badgeID` to open a dialog pretty similar to the creation dialog. IDs are not human readable, but Autocomplete will help you select the right badge.

![Screenshot from 2022-03-16 12-22-49](https://user-images.githubusercontent.com/1933730/158579272-7a7164da-0b90-412f-94f5-7a10fe5f1a1a.png)
![Screenshot from 2022-03-16 12-21-21](https://user-images.githubusercontent.com/1933730/158579256-58b3ad7b-f0c2-44f9-9d33-4679a87cd034.png)

The only difference to the creation is one extra checkbox to remove the current type or badge. If you mark this checkbox and click edit, the badge or type will be removed.
When you remove a badge, the badge is deleted permanently, along with any information about who that badge was granted to. When you remove a type, the type and all the associated badges are removed completely.

### Badge list
Badges show on several places. On the profile popover of the users, they show up to the last 20 badges granted to that user. Hovering over the badges will give you more information, and cliking on them will open the Right Hand Sidebar (RHS) with the badge details.

![Screenshot from 2022-03-16 12-29-39](https://user-images.githubusercontent.com/1933730/158580433-ca57a911-1397-432d-a739-0f06ac474845.png)

The channel header button will open the RHS with the list of all badges.

![Screenshot from 2022-03-16 12-31-18](https://user-images.githubusercontent.com/1933730/158580823-997df585-c775-43ff-9475-7a5900b151e6.png)
![Screenshot from 2022-03-16 12-32-31](https://user-images.githubusercontent.com/1933730/158580924-e24e4884-d321-465c-bd92-8c41c286612e.png)

Clicking on any badge will lead you to the badge details. Here you can check all the users that have been granted this badge.

![Screenshot from 2022-03-16 12-33-17](https://user-images.githubusercontent.com/1933730/158581085-454ff9b8-1614-4625-a4e3-16f2b0356ac8.png)

Clicking on any username on the badge details screen will lead you to the badges granted to that user.

![Screenshot from 2022-03-16 12-34-31](https://user-images.githubusercontent.com/1933730/158581257-ca614b71-3093-48fe-909d-c706c348891e.png)

## Using the Plugin API to create and grant badges
This plugin can be integrated with any other plugin in your system, to automatize the creation and granting of badges.

Using the [PluginHTTP](https://developers.mattermost.com/integrate/plugins/server/reference/#API.PluginHTTP) API method, you can create a request to the badges plugin to "Ensure" and to "Grant" the badges needed.

The badges plugin exposes the `badgesmodel` package to simplify handling several parts of this process. Some important exposed objects:
- badgesmodel.PluginPath (`/com.mattermost.badges`): The base URL for the plugin (the plugin id).
- badgesmodel.PluginAPIPath (`/papi/v1`): The plugin api route.
- badgesmodel.PluginAPIPathEnsure (`/ensure`): The ensure endpoint route.
- badgesmodel.PluginAPIPathGrant (`/grant`): The grant endpoint route.
- badgesmodel.Badge: The data model for badges.
- badgesmodel.EnsureBadgesRequest: The data model of the body of a Ensure Badges Request.
- badgesmodel.GrantBadgeRequest: The data model of the body of a Grant Badge Request.
- badgesmodel.ImageTypeEmoj (`emoji`): The emoji image type. Other image types are considered, but we recommend using emojis.

### Ensure badges
URL: `/com.mattermost.badges/papi/v1/ensure`

Method: `POST`

Body example:
```json
{
   "Badges":[
      {
         "name":"My badge",
         "description":"Awesome badge",
         "image":"smile",
         "image_type":"emoji",
         "multiple":true
      }
   ],
   "BotId":"myBotId"
}
```
Ensure badges will create badges if they already do not exist, and return the list of badges including the ids. In order to check whether a badge exist or not, it will only check the name of the badge.

### Grant badges
URL: `/com.mattermost.badges/papi/v1/grant`

Method: `POST`

Body example:
```json
{
   "BadgeID":"badgeID",
   "BotId":"myBotId",
   "UserID":"userID",
   "Reason":""
}
```
Grant badges will grant the badge with the badge id provided from the bot to the user defined. Reason is optional.
