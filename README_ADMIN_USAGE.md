# ICT Gamejam Voting Software Usage Guide

The ICT Gamejam Voting software is live at:  
https://vote.ictgamejam.com

The Administration page is at:  
https://vote.ictgamejam.com/admin

## Public Facing Page
The public page operates in two different modes:
1. Waiting
1. Voting

In 'Waiting' mode a placeholder page is displayed.  

In 'Voting' mode the voting page will be displayed. A list of games are displayed and
the user will be expected to add games to their voting ballot in their preferred order.
The ranking of the games can be adjusted after they've been added to the ballot.  
In addition to setting the Voting software to 'Voting' mode, the system that you are  
viewing the page on must be Authorized through the 'Auth Client' option in the Admin  
menu (see below) to be in `Voting` mode.

On the public facing page, hitting `Esc` will open the Admin menu.  


## Administration Page
After logging in to the Administration page the main administration page is displayed.  
The 'Public Mode' section of the page is where you switch the public page between
'Waiting' and 'Voting' modes.  
The 'Admin Sections' buttons will take you to the most used parts of Administration:
1. Votes
1. Teams
1. Games
1. Users

From the menu you can get to all parts of Adminsitration:
1. Admin - The main Admin page
1. Teams - From here you can add/edit/delete teams
1. Games - From here you can edit games
1. Votes - Here you can view all votes, along with the current voting results
1. Archive - This function doesn't actually work yet
1. Clients - From here you can view all voting clients that have been authenticated
1. Auth Client - This is used to Authorize a voting terminal
1. Users - From here you can add/edit/delete Admin Users
1. Logout - Logs you out

Most of that is self-explanatory, the most interesting part is on the 'Teams' page.  
There is a UUID listed for each team that is also a link to their Team Management page.  
Each Team can manage their own Team Members and Game information.
