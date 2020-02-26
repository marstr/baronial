# Baronial

[![Build Status](https://mstrobel.visualstudio.com/Envelopes/_apis/build/status/Baronial-CI?branchName=master)](https://mstrobel.visualstudio.com/Envelopes/_build/latest?definitionId=7?branchName=master)

Manage your personal finances with all of the power of a scriptable command-line tool, using Baronial!

## Getting Started

> Skip to [Installation Instructions](./README.md#install)

Welcome! There are thousands of personal finance and accounting tools out there, but you've stumbled onto one that
focuses on giving absolute control to you, the user. This project was inspired by [Git](https://git-scm.com), and seeks
to bring the same flexibility and power that programmers enjoy over their source code to the accounting world.

### Creating a Repository

Transaction history is captured in "repositories". Each repository holds a budget and a collection of accounts, more on
that later. To get started, open a terminal and type the following:

``` bash
$ cd
$ mkdir Finances
$ cd Finances
$ baronial init
```

You now have your first repository! There will be two empty folders in your repository "accounts" and "budget".

### Your First Transactions

#### Initialization
While Baronial does hope to be flexible, it was built with the 
[envelope system](https://en.wikipedia.org/wiki/Envelope_system) in mind. The first transaction you create will be
setting the initial state of of your accounts. Let's say you had a checking and savings account with U.S. Bank and an
Amazon credit card through Chase, with the following balances:

| Account Type | Institution | Balance  |
| :----------: | :---------: | :------: |
| Checking     | U.S. Bank   | $703.56  |
| Savings      | U.S. Bank   | $2801.22 |
| Credit Card  | Chase       | $168.91  |

You could initialize your accounts by running the following:

``` bash
$ mkdir -p accounts/us_bank/checking
$ mkdir accounts/us_bank/savings
$ mkdir -p accounts/chase/amazon
$ baronial credit 703.56 accounts/us_bank/checking
$ baronial credit 2801.22 accounts/us_bank/savings
$ baronial debit 168.91 accounts/chase/amazon
```

Notice that because the balance on your credit card is a [liability](https://www.investopedia.com/terms/l/liability.asp)
, we use the `debit` command. This is unlike the U.S. Bank accounts, which use the `credit` command, because they are 
holding [assets](https://www.investopedia.com/terms/a/asset.asp).

Now our accounts are tracking $3,335.87 in value. However, now that we know where our money is, we need to allocate it
as we intend on spending it. That's the job of the budget. When everything's working well, the sum of funds available in
the `accounts` folder should match the amount in the `budget` folder.

Let's say we pay $1,200/month in rent, we could set aside that money by running:

```bash
$ mkdir -p budget/housing/rent
$ baronial credit 1200 budget/housing/rent
```

From there, we still have $2,135.87 to work with. Let's say we put some money towards gas and groceries, and the rest in
generic savings:

```bash
$ mkdir budget/grocery
$ mkdir budget/gas
$ mkdir budget/savings
$ baronial credit 300 budget/grocery
$ baronial credit 150 budget/gas
$ baronial credit 1685.87 budget/savings
```

Now the balance of our budget should match that of our accounts! Which means that we're ready to stamp a transaction as
ready by running the following:

```bash
$ baronial commit -c "Initial account and budget balances."
```

#### Recording Expenses

Each time you spend money, you should add a transaction in the repository capturing the account and budget that it came
from. For instance, I buy coffee most mornings for $2.15 from the Café downstairs. To capture one coffee purchase, I
would type:

```bash
$ baronial debit 2.15 accounts/chase/amazon budget/grocery
$ baronial commit -m "Brewed Awakenings" -t 2019-08-01 -c "My usual morning ritual"
```

The debit command above removes $2.15 from both my credit card and my grocery budget (I know, I know, coffee from a shop
isn't exactly a grocery expense. You can categorize things however you choose.) Notice, that I can subtract the same
amount from two sources at once. This will be the most common situation, as most transactions will be categorized in a
single way. But keep in mind that if you need to split up a transaction, it isn't until the `commit` command is run that
you've finalized which budgets/accounts are impacted. You can use whatever combination of credits and debits are 
necessary to represent your transaction.

#### Income and Refunds

When funds are made available to you because of a paycheck, or maybe you've just returned an item to a store, you can
use the `credit` command to replenish your accounts/budget.

```bash
$ mkdir budget/queue
$ baronial credit 896.78 accounts/us_bank/checking budget/queue
$ baronial commit -m "MyEmployer" -t 2019-08-01 -c "First Paycheck of the Month"
```

Above, I've credited a new budget "queue" with a whole paycheck's worth of funds. I do this to even out the period in 
which I'm thinking about paychecks. For instance, I get paid twice a month, regardless of the number of days. My wife 
gets paid every two weeks, regardless of where it falls on the calendar. Trying to distribute funds from each individual
paycheck would make it hard to remember how I compared to my target each month. So instead, I accumulate money in a
buffer budget called "queue" and transfer the money to specific budgets on the first of every month. This system works 
for me, but part of the beauty of baronial is how flexible it is! Have a different way that you like to track this? Do 
it! You'll hear no complaints from me!

#### Credit Card Payments, Account Transfers

If you have credit cards, you'll be well aware of your monthly payments. Unlike other transactions mentioned above,
while money is changing hands and needs to be accounted for, you aren't actually spending money. For that reason, you're
not making a decision that impacts the budget. You're just making a decision about where your funds are, and whether you
want to be receiving or paying interest based on your circumstances.

For baronial, this is the same as a debit and a credit for the same amount against different accounts. However, it's a
common enough operation that there's a short-cure for this:

```bash
$ baronial transfer 168.91 accounts/us_bank/checking accounts/chase/amazon
$ baronial commit -m "Chase" -t 2019-08-01 -c "Monthly credit card payment"
``` 

The same command can be used to transfer money between two budgets.

### Information Recall

#### Current Balance

Now that you've created a few transactions, you need to be able to see the current status of your budgets!

The amount of funds available in each account and your top level budgets can be printed using the `balance` command.

```bash
$ baronial balance
```

Keep in mind that you can have sub-budgets inside of other budgets. You can see nested budget balances by providing the
name of the budget you'd like to see.

```bash
$ baronial balance budget/grocery
```

#### Transaction History

Want to see a list of transactions that you've input? The `log` command was made for exactly this:

```bash
$ baronial log accounts/us_bank/checking
```

Providing the extra parameter `accounts/us_bank/checking` will filter the transactions that are printed to just the ones
that touched that account. 
## Install

### Build from Source

> NOTE: To build from source, you'll need Go 1.12 or greater, perl, and Git. See [CONTRIBUTING.md](./CONTRIBUTING.md) for
more information on setting up your machine to build Baronial. 

_Unix-Based Machines:_

If you're using Linux or a Mac, take advantage of the Makefile that's included in this project. 

``` bash
git clone https://github.com/marstr/baronial.git
cd baronial
make install
```

_Windows Machines:_

It's still easy to build from source on a Windows machine.

``` Batch
git clone https://github.com/marstr/baronial.git
cd baronial
make.bat install
```