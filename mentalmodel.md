# Mental Model for this Project

This project is intuitive for SRE and people who admire the holiness of Site Reliability Engineering by Google. But for layman, developers and ops professionals the preparation of mental model for this project is a must. It is inevitable to ignore that deployments must stop strictly after some criteria are violated. This is to ensure smooth user experience and service reliability and availability. 

#### What is SLO

SLO = Service Level objective

An SLO is a promise you make about user experience, stated in measurable terms. 
Ex: 99% of requests succeeded over a 30 day window. 95% of requests finished under 300ms over a rolling week. 

**Why not just uptime?**
Uptime can answer was the service was technically reachable, but users care about did my request work, was it fast enough? Service can be up but it can return errors, time out and be partially broken. SLOs measure what users actually experience, not whether a server responded to a ping. 

> SLOs are about correctness and usefulness, not server availability.

#### What is error budget?

Error budget = 100% - SLO

Error budget is amount of failure you are allowed to have while still meeting your SLO. It can be thought of as: if SLO is 100% then failure is accepted and expected.

Ex: If your SLO is 99.9% availability over a month, your error budget is 0.1%. In a 30 day month, that means you can have about 43 minutes of downtime and still meet your SLO.

Operationally, error budget is spent when requests fail. If its gone, reliability promises are violated. 

Therefore, error budget is a control mechanism. It turns reliability into a decision making level, not a report. When the system is healthy and error budget is remaining, more frequent deployments can happen, bigger changes can be made, some instability can be accepted and feature velocity can be optimized. It means operationally, there will be fewer deployment restrictions, less fear of short-lived incidents and faster iterations. 

When error budget is exhausted, reliability becomes the top priority. Operationally, this means more cautious deployments, slower feature releases and focus on stability.


#### What is error budget remaining?

Error budget remaining simply means how much failure can be afforded. If half of error budget is gone, other half is remaining. It matters a lot for deployments, as deployments can increase risk, code changes introduce bugs, configs break things, dependencies fail. If error budget remaining is super low, it means the system is already fragile and any deployments are more likely to cause SLO violation.

Therefore, budget is nearly exhausted, deployments should be frozen. This turns reiliability into a quantitative gate, not an opinion. 


#### What does burn rate mean?

Burn rate answers question like, how fast are we spending our error budget compared to how fast we are allowed to? Its a speed comparision, not an absolute metric of error count. Consider, error budget as fuel tank. If so, burn rate is how fast the fuel is being burnt. SLO window is the race duration. If its burned too fast, SLO is violated before the window ends. 

If burn rate is 1x, it means error budget is being spent at expected rate. If its 2x, it means error budget is being spent twice as fast as allowed. If its 14x, it means error budget is being spent 14 times faster than allowed.

#### Burn rate and time (fast burn vs slow burn)

Burn rate is meaningless without time. Slow burn typically means small number of errors, which are spread out evenly, and budget will last whole SLO window. Fast burn means the contrast of it. In fact, fast burn is an emergency.

Got it. I will **continue and complete** the document, **without modifying, rewording, or correcting anything you already wrote**, and I will strictly stay in the same writing style, tone, and structure.


#### Why burn rate matters more than raw error percentage

Raw error percentage only tells how much failure has happened overall. It does not tell how urgent the situation is right now. Two systems can have the same error percentage, but one might be failing slowly over weeks, while the other might be failing rapidly in minutes.

Burn rate adds the missing dimension: time.

For example, 0.2% errors over a full 30 day window may be acceptable and within budget. But 0.2% errors in the last 5 minutes is an emergency. Burn rate captures how quickly the error budget is being consumed relative to the SLO window.

Operationally, this means burn rate tells SREs whether they should ignore the issue, investigate, page someone or block deployments immediately

This is why burn rate is more actionable than raw error percentage.

#### Mathematical formulas

**Error Budget**

```
Error Budget = 1 - SLO
```

Example:

```
SLO = 99.9% = 0.999
Error Budget = 0.001 (0.1%)
```


**Allowed Errors**

```
Allowed Errors = Total Requests × Error Budget
```


**Error Budget Remaining**

```
Error Budget Remaining =
(Allowed Errors - Observed Errors) / Allowed Errors
```

This is usually expressed as a percentage.


**Observed Error Rate**

```
Observed Error Rate = Errors / Total Requests
```


**Burn Rate**

```
Burn Rate = Observed Error Rate / Allowed Error Rate
```

Where:

```
Allowed Error Rate = Error Budget
```


#### Simple numeric example

Assume:

* SLO = 99.9%
* Error Budget = 0.1%
* Total requests in 30 days = 1,000,000

**Allowed errors**

```
1,000,000 × 0.001 = 1,000 errors
```

Now assume:

* Observed errors so far = 600

**Error budget remaining**

```
(1000 - 600) / 1000 = 0.4 = 40%
```

This means 60% of the error budget has already been consumed.


Now look at burn rate over a short window.

Last 1 hour:

* Requests = 50,000
* Errors = 100

Observed error rate:

```
100 / 50,000 = 0.002 = 0.2%
```

Allowed error rate:

```
0.1%
```

Burn rate:

```
0.2% / 0.1% = 2x
```

This means the system is consuming error budget twice as fast as allowed.

#### How these metrics are used to block deployments

Deployments introduce risk. Therefore, error budget remaining and burn rate are used as deployment gates.

Typical operational rules look like:

* If burn rate is high over a short window, deployments are blocked
* If error budget remaining falls below a threshold, deployments are frozen
* Reliability work takes priority until the system stabilizes

This removes human judgment and replaces it with objective signals. The deployment system does not care who wants to ship or how urgent a feature is. It only checks whether the system can afford more risk.

#### Common misconceptions

**“SLOs are the same as SLAs”**
SLAs are legal commitments. SLOs are engineering tools used internally to control reliability.

**“Any error is bad”**
Errors are expected and budgeted. The goal is not zero failure, but controlled failure.

**“We should aim for 100% reliability”**
100% reliability usually means zero velocity. SLOs intentionally accept failure to enable faster development.

**“If we haven’t violated the SLO yet, we’re fine”**
This ignores burn rate. You may be on track to violate the SLO very soon.

**“Burn rate is just another error metric”**
Burn rate encodes urgency. It tells how fast things are going wrong, not just that they are wrong.

