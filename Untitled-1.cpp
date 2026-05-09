#include<bits/stdc++.h>
#define print cout << "\n";
#define ll long long int
#define ull unsigned long long
#define all(v) v.begin(),v.end()
#define sz(v) ((ll)((v).size()))
//typedef vector<ll> vi;
/*#include <ext/pb_ds/assoc_container.hpp> 
#include <ext/pb_ds/tree_policy.hpp> 
using namespace std;
using namespace __gnu_pbds; 
template <typename T>
using ordered_set = tree<T, null_type,less<T>, rb_tree_tag,tree_order_statistics_node_update>;*/

using namespace std;



// SOONER OR LATER --> WE GOT EM' THE DREAM <33
// (إِنَّ اللَّهَ وَمَلَائِكَتَهُ يُصَلُّونَ عَلَى النَّبِيِّ ۚ يَا أَيُّهَا الَّذِينَ آمَنُوا صَلُّوا عَلَيْهِ وَسَلِّمُوا تَسْلِيمًا)
/*
ll check_prime(ll res)
{
	if(res==1||res==0)
		return 0;
	ll final = res;
	for(ll i = 2 ;i*i<=res;i++)
	{
		if(res%i==0)
		{
			return 0;
		}
	}
	return 1;
}

ll lcm(ll a, ll b) 
{ 
    return (a / gcd(a, b)) * b; 
} 
class students
{
	public:
	string name;
	ll A;
};

bool compare(students &a,students  &b)
{
	if(a.A==b.A)
	{
		return a.name<b.name;
	}
	return a.A>b.A;
}

ll permutation(ll n, ll k) {
    if (k > n - k) {
        k = n - k; // C(n, k) = C(n, n - k)
    }
    ll res = 1;
    for (ll i = 0; i < k; ++i) {
        res *= (n - i);
        res /= (i + 1);
    }
    return res;
}
void Mind(){
    //freopen("limited.in", "r", stdin);
    //freopen("output.txt", "w", stdout);

 
}

ll fast_power(ll base, ll power) {
    ll result = 1;
    base = base % ((ll)1e9 +7);
    while(power > 0) {

        if(power % 2 == 1) { 
            result = (result*base) %((ll)1e9 +7);
        }
        base = (base * base) %((ll)1e9 +7);
        power = power / 2; 
    }
    return result;
}
vector<ll>ar2;
vector<ll> ar(10000002);
map<ll,ll>mp_s;
void sieve()
{
	ar[0]=1;
	ar[1]=1;
	for(ll i=2;i*i<=10000001;i++)
	{
		if(ar[i]==0)
		{
			for(ll j=i*i;j<=10000001;j+=i)
			{
				if(ar[j]==0)
				{
					ar[j]=1;
				}
			}
		}
	}
	for(ll i=2;i<=10000001;i++)
	{
		if(ar[i]==0&&i%5==1)
		{
			ar2.emplace_back(i);
		}
	}
}
void prime_fact(ll n ,ll &f)
{
	vector<pair<ll,ll>>fac;
	for(ll i=2;i*i<=n;i++)
	{
		if(n%i==0)
		{
			ll c=0;
			while(n%i==0)
			{
				c++;
				n/=i;
			}
			fac.emplace_back(make_pair(i,c));
		}
	}
	if(n>1)
	{
		fac.emplace_back(make_pair(n,1));
		
	}
}

ll gcdd(ll A, ll B)
{
	if(B==0)
		return A;
	else
		return gcdd(B,A%B);
}

}*/

/*
ll permutation(ll n, ll k) {
    if (k > n - k) {
        k = n - k; // C(n, k) = C(n, n - k)
    }
    ll res = 1;
    for (ll i = 0; i < k; ++i) {
        res *= (n - i);
        res /= (i + 1);
    }
    return res;
}
*/
/*ll bin_search(vector<ll>&arr , ll ele)
{
	ll start=0,end=arr.size()-1,mid=(start+end)/2
	, in=-1;
	while(start<=end)
	{
		if(arr[mid]>=ele)
		{
			in=max(in,mid);
			start = mid+1;
		}
		else
		{
			end = mid-1;
		}
		mid=(start+end)/2;
	}
	return in+1;
}
*/
ll isprime(ll res)
{
	if(res==1||res==0)
		return 0;
	ll final = res;
	for(ll i = 2 ;i*i<=res;i++)
	{
		if(res%i==0)
		{
			return 0;
		}
	}
	return 1;
}
void solve()
{
	ll n,k;
	cin >> n >> k;
	vector<ll>ar(n+1);
	for(ll i=0;i<n;i++)
		cin >> ar[i];
	map<ll,multiset<ll>>mp;
	for(ll i=0;i<n;i++)
	{
		mp[ar[i]%k].insert(ar[i]);
	}
	ll ans=0;
	for(ll i=0;i<n;i++)
	{
		ll nu = ar[i]%k;
		ll com = k-nu;
		if(com==k)
			com=0;
		if(nu==com && mp[nu].size()>=2)
		{
			ans+=mp[nu].size()-1;
			auto it = mp[nu].find(ar[i]);
			mp[nu].erase(it);
		}
		else if(nu!=com && mp[nu].find(ar[i])!=mp[nu].end() && mp[com].size()>=1)
		{
			ans+=mp[com].size();
			auto it = mp[nu].find(ar[i]);
			mp[nu].erase(it);
		}
	}
	cout << ans;
	
}
int main()
{
    ios_base::sync_with_stdio(false), cin.tie(NULL), cout.tie(NULL);
    //sieve();
    //Mind();
    ll t=1;
    //cin >> t;
    while(t--)
    {
        solve();
        print;
    }
}